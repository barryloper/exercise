package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

// Command-line flags
// todo: allow these to be positional
var (
	port    = flag.Int("port", 8000, "Enter a port number on which the server will listen")
	address = flag.String("address", "localhost", "Enter the FQDN or IP address on which the server will listen")
)


func main() {
	flag.Parse()

	// combine port and address flags
	listenURL := fmt.Sprintf("%s:%d", *address, *port)
	http.HandleFunc("/", DocsHandler)
	http.HandleFunc("/hash", HashHandler)
	http.HandleFunc("/stats", StatsHandler)
	//todo: allow port as cli arg
	fmt.Printf("Listening on %s\n", listenURL)
	log.Fatal(http.ListenAndServe(listenURL, nil))
}
