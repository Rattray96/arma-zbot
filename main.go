package main

import (
	"log"

	"github.com/zanven42/arma-zbot/zanbot"
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func main() {
	zanbot.Init()
}
