package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
func TestAddHash(t *testing.T) {
	t.Parallel()
	const numHashesToTest int = 100
	const maxPasswordLengthBytes int = 64
	baseHashTable := &PasswordStore{}

	var wg sync.WaitGroup
	wg.Add(numHashesToTest)
	for i := 0; i < numHashesToTest; i++ {

		go func() {
			defer wg.Done()
			passwordLength := rand.Intn(maxPasswordLengthBytes)
			password := make([]byte, passwordLength)
			rand.Read(password)
			newHashID, resultChannel, err := baseHashTable.addHash(password)
			if err != nil {
				t.Error("Error beginning the hash add")
				t.Fail()
			}
			<-resultChannel //wait for result

			if !baseHashTable.checkPassword(newHashID, password) {
				t.Errorf("Password %s didn't match for user %d", string(password), newHashID)
				t.Fail()
			}
		}()

	}
	wg.Wait()

	t.Log(len(baseHashTable.passwords), " hashes computed")
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
