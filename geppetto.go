package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/forklift/geppetto/engine"
	"github.com/forklift/geppetto/unit"
)

var port = flag.String("port", "5000", "Define what TCP port to bind to")
var base = flag.String("base", "/etc/geppetto", "Service files location.")

var Engine *engine.Engine

func main() {

	flag.Parse()
	endpoint := ":" + *port
	unit.BasePath = *base

	Engine = engine.New()

	go func() {
		for e := range Engine.Events {
			fmt.Println(e)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/_ping", pong)
	mux.HandleFunc("/start", start)

	log.Printf("Listening at %s", endpoint)
	log.Fatal(http.ListenAndServe(endpoint, mux))
}

func pong(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}

func start(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Request.")
	w.Write([]byte("Hello.\n"))
	name := r.FormValue("start")

	if name == "" {
		w.WriteHeader(400)
		w.Write([]byte("Missing Start Value"))
		return
	}

	units := unit.Make([]string{name})

	for e := range Engine.Start(units...) {
		w.Write([]byte(e.String()))
	}
	w.Write([]byte("\nDone."))
}
