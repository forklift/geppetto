package main

import (
	"flag"
	"log"
	"net/http"
)

var port = flag.String("port", "5000", "Define what TCP port to bind to")

func main() {

	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/_ping", pong)

	endpoint := ":" + *port
	log.Printf("Listening at %s", endpoint)
	log.Fatal(http.ListenAndServe(endpoint, mux))
}

func pong(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}
