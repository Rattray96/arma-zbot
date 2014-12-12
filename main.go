package main

import (
	"log"

	"github.com/zanven42/arma-zbot/zanbot"

	"os"
	"time"
	//"os/exec"
	"io"
	//"bytes"
	//"bufio"
	//"strings"
	//"fmt"
	"regexp"
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func main() {

	zanbot.Init()
	//buffer := bytes.NewBuffer([]byte(""))

	fileSize := int64(0)

	for {
		<-time.After(1 * time.Second)
		fi, _ := os.Stat("./src/remoteexec.log")
		sz := fi.Size()

		var b []byte
		switch {
		case sz < fileSize:
			// reread entire file should be very rare case.
			b = make([]byte, sz)
			fileSize = 0
		case sz > fileSize:
			// read from old to new
			b = make([]byte, sz-fileSize)
		case sz == fileSize:
			// nothing changed who gives a shit.
			continue
		}
		file, err := os.OpenFile("./src/remoteexec.log", os.O_RDONLY, 0444)
		if err != nil && err != io.EOF {
			panic(err)
		}
		file.ReadAt(b, fileSize)
		err = file.Close()
		if err != nil {
			panic(err)
		}
		s := string(regexp.MustCompile("[^\"\n]\n\t").ReplaceAll(b, []byte("")))
		println(s)
		// set filesize to new save point.
		fileSize = sz
	}
}
