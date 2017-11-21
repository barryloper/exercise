package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
}
func TestAddHash(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	const numHashesToTest int = 50000
	wg.Add(numHashesToTest)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	baseHashTable := &HashTable{}

	for i := 0; i < numHashesToTest; i++ {

		go func() {
			defer wg.Done()
			passwordLength := r.Intn(64)
			password := make([]byte, passwordLength)
			rand.Read(password)
			//fmt.Println("baseHashTable  ", &baseHashTable)
			newHashID := baseHashTable.addHash(password)
			time.Sleep(8 * time.Second) // can we get something to wait on here?
			if !baseHashTable.checkPassword(newHashID, password) {
				t.Errorf("Password %s didn't match for user %d", string(password), newHashID)
			}
		}()

	}

	wg.Wait()

	t.Log(len(baseHashTable.table), " hashes computed")
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
