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
}
