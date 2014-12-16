package zanbot

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type rconBanInfo struct {
	id       string
	guid     string
	duration string
	reason   string
}

func rconManager(ch chan banData, rl chan []rconinfo, testMode bool) {
	mch := []chan banData{}
	f, err := os.OpenFile("./zbot-Log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("cant open log file. ./zbot-Log.txt")
	}
	defer f.Close()
	log.SetOutput(f)

	for {
		select {
		case b := <-ch:
			// fan out ban to all rcon consoles. aint this cool aye :D.
			for i := 0; i < len(mch); i++ {
				mch[i] <- b
			}
			// first load or when the settings file has been moded.
		case nrconl := <-rl:
			// close all rcons to reboot with new rcon settings
			for i := 0; i < len(mch); i++ {
				close(mch[i])
			}
			mch = []chan banData{}
			for i := 0; i < len(nrconl); i++ {
				c := make(chan banData)
				go rcon(nrconl[i], c, testMode)
				mch = append(mch, c)
			}
			// reload settings and relaunch consoles
		}
	}

}
func rcon(rc rconinfo, ch chan banData, testing bool) {
	// bans the client with the given rcon

	//setup rcon connection and make sure it stays connected

	fmt.Println("starting rcon @ host:", rc.Host)
	cmd := exec.Command("BattleNETclient.exe", "-host", rc.Host, "-port", rc.Port, "-password", rc.Pw)
	// set output pipes to pipe info from battlenet's output
	ro, wo := io.Pipe()
	cmd.Stdout = wo
	scanner := bufio.NewScanner(ro)
	// setup input pipes to send info to battlenet
	ri, wi := io.Pipe()
	cmd.Stdin = ri
	cmd.Start()

	rconch := make(chan string)
	ech := make(chan bool)

	go func(ch chan bool, cmd *exec.Cmd) {
		cmd.Wait()
		ch <- true
	}(ech, cmd)

	go func(s chan string, scanner *bufio.Scanner) {
		for scanner.Scan() {
			s <- scanner.Text()
		}
	}(rconch, scanner)

	// store a array of banned people to help mitigate double banning
	var baned []rconBanInfo
	baning := make(map[string]banData)
	// start a 10min timer which is used for validating bans have occured and for removing duplicate bans that may have happened.
	// this loop solves the issue of banning as rcon gets reset as it will fix itself within 10mins.
	tba := time.After(10 * time.Second)
	for {
		select {
		// do a banlist validation. to make sure were doing our job
		case <-tba:
			// reset the timer
			tba = time.After(1 * time.Minute)
			wi.Write([]byte("bans\n"))

		// rcon has closed reboot that motherfucker.
		case <-ech:
			// small delay so we dont dos the server
			time.After(1 * time.Second)
			println("restarting rcon:", rc.Host)
			cmd = exec.Command("BattleNETclient.exe", "-host", rc.Host, "-port", rc.Port, "-password", rc.Pw)
			// set output pipes to pipe info from battlenet's output
			ro, wo = io.Pipe()
			cmd.Stdout = wo
			scanner = bufio.NewScanner(ro)
			// setup input pipes to send info to battlenet
			ri, wi = io.Pipe()
			cmd.Stdin = ri
			cmd.Start()
			go func(ch chan bool, cmd *exec.Cmd) {
				cmd.Wait()
				ch <- true
			}(ech, cmd)
			go func(s chan string, scanner *bufio.Scanner) {
				for scanner.Scan() {
					s <- scanner.Text()
				}
			}(rconch, scanner)
		//ban a user and make sure we get a "banned response" from rcon before procedding
		case b, ok := <-ch:
			// if the channel was closed time to return and let it reconnect.
			if !ok {
				return
			}

			// if he is already on the banning list then we have already tried to ban him so dont dup the bans (if he hasnt actuall been banned it will be cleaned up shortly.)
			if baning[b.guid].guid != "" {
				continue
			}
			// on first bootup the entire file will be searched so dont add a duplicate ban wait till we recieve the banlist and this list will be sorted on its own later
			if len(baned) == 0 {
				println("Adding Ban Check for:", b.name, "- Guid:", b.guid)
				baning[b.guid] = b
				continue
			}

			// if he has definetly already been banned then fuck what we still doing in here just end. (this check means we can abuse this and pass banlists between rcons and sync them)
			for i := 0; i < len(baned); i++ {
				if baned[i].guid == b.guid {
					continue
				}
			}
			if testing {
				log.Println("TestMode:", rc.Host, "| Banned:", b.name, "-", b.guid, "File:", b.file, "Reason:", b.reason)
				continue
			}
			command := "addban " + b.guid + " 0 " + b.name + " | Hacking | PERM | zBot"
			if b.file == "Web" {
				command = "addban " + b.guid + " 0 " + b.reason
			}
			wi.Write([]byte(command + "\n"))
			baning[b.guid] = b
			println("Baning:", b.name, "-", b.guid)
			//tn := time.Now()
			log.Println(rc.Host, "| Banned:", b.name, "-", b.guid, "File:", b.file, "Reason:", b.reason)
			// add the ban to our list to be checked next time the banlist is scanned. to verify he actually got banned.

		case s := <-rconch:
			// we scan the console if nothing is ready and break after each line to see if something is ready to be done.
			// primary handler of disconnects anything that consumes the scanner needs to take this job over while consuming.
			txt := s
			if txt == "GUID Bans:" {
				counter := 0
				//baned = make([]rconBanInfo,0)
				// consume the scanner till we have all the bans we are after
				baned = make([]rconBanInfo, 0)
				for scanner.Scan() {
					btxt := scanner.Text()
					// skip markup code
					if btxt == "[#] [GUID] [Minutes left] [Reason]" || btxt == "----------------------------------------" {
						continue
					}
					// if we are at the bottom its a blank line so this will break us out of the loop once we hit the bottom
					if len(btxt) == 0 {
						break
					}
					counter++
					// split txt into a slice of words
					words := strings.Fields(btxt)
					// if reason is more then one word concat it down to one element in the array
					if len(words) < 4 {
						// if we dont have a reason then we dont want to handle the ban.
						continue
					}

					if len(words) > 4 {
						for i := 4; i < len(words); i++ {
							words[3] = words[3] + " " + words[i]
						}
						words = words[0:4]
					}

					var rcb rconBanInfo
					rcb.id = words[0]
					rcb.guid = words[1]
					rcb.duration = words[2]
					rcb.reason = words[3]
					baned = append(baned, rcb)
				}

				// check for duplicate bans and unban them.
				// start from highest number. if duplicate found at a lower number remove self.
				unbanverify := make(map[string]int)
				for i := len(baned) - 1; i > 0; i-- {
					// j should always be less then i. because anything above it should have already been checked.
					for j := 0; j < i; j++ {
						if baned[i].guid == baned[j].guid {
							if baned[i].duration == "perm" && baned[j].duration == "perm" {
								//
								// this whole check has me bedazled as to why since im sure i dont get duplicate ban ids from rcon. but in practice i do for some fucked up reason
								if unbanverify[baned[i].id] == 0 {
									unbanverify[baned[i].id] = 1
									if testing {
										println("TestMode: removed duplicate ban:", baned[i].guid)
										log.Println("TestMode: Removed Duplicate ban id:", baned[i].id, "- guid:", baned[i].guid)
										continue
									}
									println("removed duplicate ban:", baned[i].guid)
									log.Println("Removed Duplicate ban id:", baned[i].id, "- guid:", baned[i].guid)
									wi.Write([]byte("removeBan " + baned[i].id + "\n"))
									<-time.After(100 * time.Millisecond)
								}
							}
						}
					}
				}
				// ok now check if the people we just banned are on the ban list. if they are remove
				for i := 0; i < len(baned); i++ {
					// safe delete. does nothing by default if doesnt exist in the map.
					delete(baning, baned[i].guid)
				}
				// send whoever is left back to this rcon to be rebanned on next select.
				for _, v := range baning {
					// do in a gofunc so its non blocking. so we can do each and not deadlock.
					go func(bdata banData, ch chan banData) {
						println(bdata.name, "- was not banned is being banned now")
						ch <- bdata
					}(v, ch)
				}
				// remake the baning map since we have cleared it or pushing all the bans back to be rebanned.
				baning = make(map[string]banData)
				continue
			}
			// check if we cant connect or lost connection.
			if txt == "Host unreachable!" {
				cmd.Process.Kill()
				println("killing process")
			}
			// consume an extra
			if txt == "Disconnected! (Connection timeout)" {
				// consume an extra scan because the next message will be "Connected!"
				println("Rcon connection Lost with:", rc.Host, "Attempting Reconnect")
				scanner.Scan()
				t := scanner.Text()
				if t == "Connected!" {
					println("Rcon Reconnected with:", rc.Host)
				}
			}
			//fmt.Println(txt)
		}
	}
}
