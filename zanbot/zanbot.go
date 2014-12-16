package zanbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
	//"os/exec"
	"bufio"
	//"io"
	"log"
)

var gbanchan chan banData

//var grlchan chan []rconinfo
var tLoc *time.Location

//var tStart time.Time
func main() {
	Init()
}

// Init Starts the Bot
func Init() {
	go listen()
	var settings jsonsettings
	fmt.Println("zBot Created By Anthony \"Zanven\" Poschen.")

	gbanchan = make(chan banData)
	rlchan := make(chan []rconinfo)
	settingfile := "./zbot-settings.json"
	defer close(gbanchan)

	var err error
	tLoc, err = time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}

	fs, err := os.Stat(settingfile)
	if err != nil {
		panic(err)
	}
	tSSF := fs.ModTime()

	sf, err := ioutil.ReadFile(settingfile)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(sf, &settings)
	go rconManager(gbanchan, rlchan, settings.TestMode)
	rlchan <- settings.Rcons

	// @TODO: convert this into its own file manager and pass in new configs if the file changes so this gets hot reloaded if its valid and closes previous.
	_, err = os.Stat(settings.RemotExecFile)
	if err == nil {
		go filemonitor(settings.RemotExecFile, RemoteExecFile)
	}
	_, err = os.Stat(settings.PublicVarFile)
	if err == nil {
		go filemonitor(settings.PublicVarFile, PublicVarFile)
	}
	if err == nil {
		go filemonitor(settings.ScriptsFile, ScriptsFile)
	}

	//processfile(lines,tSRE)
	for {
		// dont spam the cpu to do file i/o every frame. so idle a reasonable ammount between loops.
		<-time.After(1000 * time.Millisecond)

		f, err := os.Stat(settingfile)
		if err != nil {
			continue
		}
		t := f.ModTime()
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, tLoc)

		if tSSF.Before(t) {
			println("Reloading Settings")
			time.Sleep(100 * time.Microsecond)
			sf, err := ioutil.ReadFile(settingfile)
			if err != nil {
				panic(err)
			}
			settings = jsonsettings{}
			json.Unmarshal(sf, &settings)
			rlchan <- settings.Rcons
			tSSF = t
		}
	}

}

func filemonitor(filename string, fn func(*bufio.Scanner)) {
	t := time.After(100 * time.Millisecond)
	offset := int64(0)
	st := time.UTC
	st.String()
	// infite loop cause this should never be changed
	for {
		<-t
		t = time.After(100 * time.Millisecond)
		fi, err := os.Stat(filename)
		if err != nil {
			panic(err)
		}

		sz := fi.Size()
		//var b []byte
		switch {
		//send entire file. if required or if file is now shorter then it use to be.
		// this should only occur if the file is manually modified to be shorter. or if its the settings file since we want the entire file.
		case sz < offset:
			//b = make([]byte, sz)
			offset = 0
		// a new entry has been added delta read the new part
		case sz > offset:
			//b = make([]byte, sz-offset)
		case sz == offset:
			continue
		}
		// @TODO: add a delay to reading file if fail to open.
		file, err := os.OpenFile(filename, os.O_RDONLY, 0444)
		if err != nil /*&& err != io.EOF*/ {
			log.Fatal(err)
			continue
		}
		file.Seek(offset, os.SEEK_SET)
		offset = sz
		sc := bufio.NewScanner(file)

		// send file data to processor.
		fn(sc)
		// dont launch func in go routine cause closing the file exits the function pre maturarly. so just wait for it to return.
		file.Close()
	}

}

func validateLine(s string, t int) int {
	r := 0
	l := len(s) - 1
	if l < 0 {
		l = 0
	}
	re, _ := regexp.Compile(`^\d\d\.\d\d\.\d\d\d\d \d\d:\d\d:\d\d:.*`)
	switch t {
	case 0:
		//check if its a start of a new entry

		if (string(s[l]) == string("\"") && l > 71) && re.Match([]byte(s)) {
			if l != strings.Index(s, "\"") {
				r = 1
			}
		}
	case 1:
		if l == 0 {
			return 0
		}
		if (string(s[l]) == string("]") && l > 71) && re.Match([]byte(s)) {
			if l != strings.Index(s, "]") {
				r = 1
			}
		}
	case 2:
		if l > 140 && re.Match([]byte(s)) {
			r = 2
			if l != strings.Index(s, "\"") {
				r = 1
			}

		}

	}
	return r
}

func getline(sc *bufio.Scanner, t int) ([]string, bool) {
	var lines []string
	if sc.Scan() {
		txt := sc.Text()
		if validateLine(txt, t) == 1 {
			lines = append(lines, txt)
			return lines, true
		}
		// if not valide we now have to get a valid line and have to get the following line after it and also return that to avoid skipping lines
		// for loop to make sure we dont stop till all conditions met
		for {
			if !sc.Scan() {
				break
			}
			s := sc.Text()
			vl := validateLine(s, t)
			if vl == 1 {
				lines = append(lines, txt)
				lines = append(lines, s)
				return lines, true
			}
			if vl == 2 {
				lines = append(lines, txt)
				txt = ""
			}
			txt = txt + s
		}
	}
	return lines, false
}

// RemoteExecFile calls remote exec
func RemoteExecFile(sc *bufio.Scanner) {
	for {
		lines, b := getline(sc, 0)
		if !b {
			break
		}
		//println(txt)
		for i := 0; i < len(lines); i++ {
			pos := strings.Index(lines[i], " - #0 \"")
			if pos == -1 {
				pos = strings.Index(lines[i], " - Compile Block")
			}
			// check if its the ignore case we dont want to check for
			r := strings.Contains(lines[i][pos+6:], "\"removeBackpack this; removeAllWeapons this;\"")
			if !r {
				go ban(lines[i], "RemoteExec")
			}
		}
	}

}

// PublicVarFile called pub file
func PublicVarFile(sc *bufio.Scanner) {
	re1, _ := regexp.Compile(`.*=.*=.*=`)
	re2, _ := regexp.Compile(`compile|compileFinal|parseText|toString`)
	for {

		lines, b := getline(sc, 1)
		if !b {
			break
		}
		for i := 0; i < len(lines); i++ {
			b := []byte(lines[i])
			if re1.Match(b) {
				if re2.Match(b) {
					go ban(lines[i], "PublicVar")
				}
			}
		}
	}
}

// ScriptsFile finds all matches to current only case thats unique. needs improvement. or better bec filters to detect more shit.
func ScriptsFile(sc *bufio.Scanner) {
	re, _ := regexp.Compile("(?i)(removeallactions|RscDisplayGame)")
	for {
		lines, b := getline(sc, 2)
		if !b {
			break
		}
		for i := 0; i < len(lines); i++ {
			b := []byte(lines[i])
			if re.Match(b) {
				//go ban(lines[i], "Scripts")
				println(len(lines), i, lines[i])
			}
		}

	}
	println("finished")
}

func ban(s string, f string) {
	name := getPlayerName(s)
	guid := getPlayerGUID(s)
	reason := getBanReason(s)
	gbanchan <- banData{name, guid, reason, f}
}
