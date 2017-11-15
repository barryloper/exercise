package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestDocsHandler(t *testing.T) {}

//A GET to /stats should accept no data;
//it should return a JSON data structure for the total hash requests since server start
//and the average time of a hash request in milliseconds.
func TestGetStats(t *testing.T) {
	req:= httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	StatsHandler(w, req)
	fmt.Printf("got %q\n", w.Body.String())
}

//A POST to /hash should accept a password;
//it should return a job identifier immediate;
//it should then wait 5 seconds and compute the password hash.
//The hashing algorithm should be SHA512.
func TestPostHash(t *testing.T) {}

//A GET to /hash should accept a job identifier;
//it should return the base64 encoded password hash for the corresponding POST request.
func TestGetHash(t *testing.T) {}
