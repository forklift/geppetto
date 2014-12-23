package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/forklift/operator/engine"
	"github.com/forklift/operator/unit"
)

var (
	host     = flag.String("host", ":9090", "Define what TCP port to bind to")
	base     = flag.String("path", "/etc/operator", "Geppetto files base path.")
	start    = flag.String("start", "", "List of units to start by defualt.")
	insecure = flag.Bool("insecure", false, "Don't use TLS for communications.")
)

var Engine *engine.Engine

func main() {

	flag.Parse()

	unit.BasePath = path.Join(*base, "services")

	Engine = engine.New()
	//	engineLog := make(chan event.Event)
	//	Engine.Listeners.Add("logs", engineLog)

	mux := http.NewServeMux()

	mux.HandleFunc("/_ping", APIpong)
	mux.HandleFunc("/start", APIstart)
	mux.HandleFunc("/stop", APIstop)

	var wg sync.WaitGroup

	//Log the events.
	wg.Add(1)
	go func() {
		defer wg.Done()
		//for e := range engineLog {
		//	log.Println(e)
		//}
	}()

	//Start the server.
	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("Listening at %s", *host)
		var err error
		if *insecure {
			err = http.ListenAndServe(*host, mux)
		} else {
			err = http.ListenAndServeTLS(*host, path.Join(*base, "cert.pem"), path.Join(*base, "key.pem"), mux)
		}

		log.Fatal(err)

	}()

	//Start the services...
	wg.Add(1)
	go func() {
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

func APIpong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
	log.Printf("ping from %s (%s)\n", r.RemoteAddr, r.UserAgent())
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
