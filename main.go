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
	// initialize the data store
	db := NewPasswordStore()
	// pass the data store to the configured muxer
	// configures and handles the routes (handlers.go)
	muxer := MakeMuxer(db)
	// combine port and address cli flags
	listenURL := fmt.Sprintf("%s:%d", *address, *port)

	log.Printf("Listening on %s\n", listenURL)
	log.Fatal(http.ListenAndServe(listenURL, muxer))

}
