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
func TestPasswordStore_updateAverageHashTime(t *testing.T) {
	// Tests that updating the average asynchronously results in the expected average

	t.Parallel()

	store := NewPasswordStore()
	var times [100]time.Duration //an array of integers representing seconds
	var sumtimes time.Duration
	for i := range times {
		times[i] = time.Duration(rand.Int63n(24)) * time.Hour
		sumtimes += times[i]

	}
	expectedAverage := sumtimes / time.Duration(len(times)) // expected average time

	var wg sync.WaitGroup
	wg.Add(len(times))
	for _, t := range times {
		// updates the average time with a bunch of goroutines
		go func(t time.Duration) {
			defer wg.Done()
			store.updateAverageHashTime(time.Duration(t))
		}(t)
	}
	wg.Wait()

	if store.averageHashTime-expectedAverage > time.Duration(1)*time.Millisecond {
		t.Log("Wanted average to be ", expectedAverage, " got ", store.averageHashTime)
		t.Fail()
	}

}

func TestAddHash(t *testing.T) {
	t.Parallel()
	const numHashesToTest int = 100
	const maxPasswordLengthBytes int = 64
	baseHashTable := NewPasswordStore()

	wg := sync.WaitGroup{}
	for i := 0; i < numHashesToTest; i++ {

		wg.Add(1)
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

			if !baseHashTable.CheckPassword(newHashID, password) {
				t.Errorf("Password %s didn't match for user %d", string(password), newHashID)
				t.Fail()
			}
		}()

	}
	baseHashTable.Sync()
	wg.Wait()

	count, duration := baseHashTable.GetStats()
	t.Log("Added", count, "hashes in", duration)
}
