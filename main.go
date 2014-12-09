package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/forklift/geppetto/engine"
	"github.com/forklift/geppetto/unit"
)

var port = flag.String("port", "5000", "Define what TCP port to bind to")
var units = flag.String("units", "/etc/geppetto", "Units files path.")

var Engine *engine.Engine

func main() {

	flag.Parse()
	endpoint := ":" + *port
	unit.BasePath = *units

	Engine = engine.New()

	go func() {
		for e := range Engine.Events {
			fmt.Println(e)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/_ping", pong)
	mux.HandleFunc("/start", start)
	mux.HandleFunc("/stop", stop)

	log.Printf("Listening at %s", endpoint)
	log.Fatal(http.ListenAndServe(endpoint, mux))
}

func pong(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}

func start(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Request.")

	defer w.Write([]byte("\nDone."))

	names := strings.Split(r.FormValue("units"), ",")

	if len(names) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Error: No Units to start."))
		return
	}

	for _, name := range names {
		for e := range Engine.Start(name) {
			w.Write([]byte(e.String()))
		}
	}

	fmt.Println("Request End.")
}

func stop(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Request.")

	defer w.Write([]byte("\nDone."))

	names := strings.Split(r.FormValue("units"), ",")

	if len(names) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Error: No Units to start."))
		return
	}

	for _, name := range names {
		for e := range Engine.Stop(name) {
			w.Write([]byte(e.String()))
		}
	}

	fmt.Println("Request End.")
}

/*
func signal(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Request.")

	defer w.Write([]byte("\nDone."))

	names := strings.Split(r.FormValue("units"), ",")

	var err []byte

	if len(names) == 0 {
		err = []byte("Error: No Units to start.")
	}

	sig, ok := signals[r.FormValue("sig")]

	if !ok {
		err = []byte("Unsupported signal.")
	}

	if err != nil {
		w.WriteHeader(400)
		w.Write(err)
		return
	}

	units := unit.Make(names)

	for _, u := range units {
		for e := range Engine.Signal(u) {
			w.Write([]byte(e.String()))
		}
	}

	fmt.Println("Request End.")
}

*/
