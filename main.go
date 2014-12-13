package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/forklift/geppetto/engine"
	"github.com/forklift/geppetto/unit"
)

var (
	host  = flag.String("host", ":9090", "Define what TCP port to bind to")
	path  = flag.String("path", "/etc/geppetto", "Units files path.")
	start = flag.String("start", "", "List of units to start by defualt.")
)

var Engine *engine.Engine

func main() {

	flag.Parse()
	unit.BasePath = *path

	Engine = engine.New()

	mux := http.NewServeMux()

	mux.HandleFunc("/_ping", APIpong)
	mux.HandleFunc("/start", APIstart)
	mux.HandleFunc("/stop", APIstop)

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()

		for e := range Engine.Events {
			fmt.Println(e)
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()

		log.Printf("Listening at %s", *host)
		log.Fatal(http.ListenAndServe(*host, mux))
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		if *start == "" {
			return
		}

		log.Println("Starting services...")
		for _, name := range strings.Split(*start, " ") {
			for e := range Engine.Start(name) {
				log.Println([]byte(e.String()))
			}
		}
	}()

	wg.Wait()
	log.Println("Exiting.")
}

func APIpong(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}

func APIstart(w http.ResponseWriter, r *http.Request) {

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

func APIstop(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Request.")

	defer w.Write([]byte("\nDone."))

	names := strings.Split(r.FormValue("units"), ",")

	if len(names) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Error: No Units to start."))
		return
	}

	w.Write([]byte("Hello"))
	for _, name := range names {
		for e := range Engine.Stop(name) {
			w.Write([]byte(e.String()))
		}
	}

	fmt.Println("Request End.")
}
