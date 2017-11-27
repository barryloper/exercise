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
	// initialize the store
	// pass the store into a custom handler that implements http, or custom server?
	// function that returns a handler function?
	// or pass a channel into those things so they can communicate with
	// a data-store goroutine...

	db := NewPasswordStore()
	muxer := MakeMuxer(db)
	// combine port and address flags
	listenURL := fmt.Sprintf("%s:%d", *address, *port)

	//muxer.HandleFunc("/hash/", RestEndpoint("/hash/", MethodMap{http.MethodGet: HashHandler}))

	fmt.Printf("Listening on %s\n", listenURL)
	log.Fatal(http.ListenAndServe(listenURL, muxer))
}
