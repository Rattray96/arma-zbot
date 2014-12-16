package zanbot

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// need to setup https on server if donations pay for a cert.
// need to setup a secure and prefably simple system to handle authentication.
// probs go down the route of building certs server side and having the users download them and place them in a certain spot for zbot to read and use them.
// to allow the server to securely communicate with the client and vise versa. although that method has issues and security flaws best
// i can think of off the top of my head.

type guid struct {
	GUID string
}

func listen() {
	http.HandleFunc("/api/addban", addbanHandle)
	http.HandleFunc("/api/unban", unbanHandle)
	http.HandleFunc("/api/bansync", bansyncHandle)

	http.ListenAndServe(":4000", nil)

}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Incoming msg")

	w.Write([]byte(r.Header.Get("")))
}

func addbanHandle(w http.ResponseWriter, r *http.Request) {
	var t guid
	err := decodeJson(r, &t)
	if err != nil || t.GUID == "" {
		errorHandler(w, r, 500)
		return
	}
	fmt.Println("Recieved Ban Request for:", t.GUID)
	// do a bunch of authentication stuff

	//add bancheck if authentication pass and it isnt malicious request.
	// do it in a go func so non blocking and can return status code
	go func(s string) {
		gbanchan <- banData{"", s, "zbot | Community | PERM", "Web"}
	}(t.GUID)
}

func unbanHandle(w http.ResponseWriter, r *http.Request) {
	var t guid
	err := decodeJson(r, &t)
	if err != nil || t.GUID == "" {
		errorHandler(w, r, 500)
		return
	}
	fmt.Println("Recieved Unban Request for:", t.GUID)
	go func(s string) {
		// do some stuff to unban
	}(t.GUID)
}

func bansyncHandle(w http.ResponseWriter, r *http.Request) {
	// unban and ban people acordingly.
}

func decodeJson(r *http.Request, t interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&t)
	return err
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
}
