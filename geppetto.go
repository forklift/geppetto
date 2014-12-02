package main

import (
	"log"
	"net/http"
)

var rh = http.RedirectHandler("http://example.org", 307)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/_ping", pong)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":5000", mux))
}

func pong(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}
