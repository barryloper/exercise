package main

import (
	"testing"
)

//When launched it should monitor a given port and wait for http connections
func TestListening(t *testing.T) {}

//A POST to /hash should accept a password;
//it should return a job identifier immediate;
//it should then wait 5 seconds and compute the password hash.
//The hashing algorithm should be SHA512.
func TestPostHash(t *testing.T) {}

//A GET to /hash should accept a job identifier;
//it should return the base64 encoded password hash for the corresponding POST request.
func TestGetHash(t *testing.T) {}

//A GET to /stats should accept no data;
//it should return a JSON data structure for the total hash requests since server start
//and the average time of a hash request in milliseconds.
func TestGetStats(t *testing.T) {}

//The software should be able to process multiple connections simultaneously
func TestSimultaneousRequest(t *testing.T) {}

//The software should support a graceful shutdown request.
//it should allow any remaining password hashing to complete,
//reject any new requests, and shutdown.
func TestGracefulShutdown(t, *testing.T) {}

//No additional password requests should be allowed when shutdown is pending.
func TestShuttingDownPasswordRequest(t, *testing.T) {}
