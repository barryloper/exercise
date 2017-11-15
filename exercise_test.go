package main

import (
	"fmt"
	"testing"
)

func init() {
	fmt.Println("Initialized!")
}

//When launched it should monitor a given port and wait for http connections
func TestListening(t *testing.T) {}

//The software should be able to process multiple connections simultaneously
func TestSimultaneousRequest(t *testing.T) {}

//The software should support a graceful shutdown request.
//it should allow any remaining password hashing to complete,
//reject any new requests, and shutdown.
func TestGracefulShutdown(t *testing.T) {}

//No additional password requests should be allowed when shutdown is pending.
func TestShuttingDownPasswordRequest(t *testing.T) {}
