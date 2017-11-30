package main

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestAddHash(t *testing.T) {
	hashTable := NewPasswordStore()
	const requestTimeout time.Duration = 7 * time.Second

	testFn := func(password []byte, t *testing.T) {
		t.Parallel()

		_, response := hashTable.SavePassword(password)
		select {
		case <-response:
			return
		case <-time.After(requestTimeout):
			t.Fatal("Timeout waiting for hash response")
		}
		return
	}
	for i := 0; i < numHashesToTest; i++ {
		password := RandomPassword(minPasswordLength, maxPasswordLength)
		t.Run(string(password), func(t *testing.T) { testFn(password, t) })

	}
	stats := hashTable.GetStats()
	t.Log("Added", stats.Total, "hashes. Avg time:", stats.Average)
}
