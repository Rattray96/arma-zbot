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

	//go listen()
	zanbot.Init()

}

/*
func listen() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":4000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Incoming msg")
	w.Write([]byte(r.))
}
*/
