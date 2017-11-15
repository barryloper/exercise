package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Command-line flags
// todo: allow these to be positional
var (
	port    = flag.Int("port", 8000, "Enter a port number on which the server will listen")
	address = flag.String("address", "localhost", "Enter the FQDN or IP address on which the server will listen")
)

type hashServer struct {
	data *map[int]struct {
		hash        string
		computeTime int
	}
	dataLock      *sync.RWMutex
	averageTime   *int
	timeLock      *sync.RWMutex
	totalRequests *int
	totalLock     *sync.RWMutex
}

func (h hashServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// do stuff
	w.WriteHeader(404)
	return
}

func main() {
	flag.Parse()

	listener := &hashServer{}
	// combine port and address flags
	listenURL := fmt.Sprintf("%s:%d", *address, *port)
	http.HandleFunc("/", DocsHandler)
	http.HandleFunc("/hash", HashHandler)
	http.HandleFunc("/stats", StatsHandler)
	//todo: allow port as cli arg
	fmt.Printf("Listening on %s\n", listenURL)
	log.Fatal(http.ListenAndServe(listenURL, *listener))
}
