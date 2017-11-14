package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hash", HashHandler)
	http.HandleFunc("/stats", StatsHandler)
	//todo: allow port as cli arg
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", nil))
}
