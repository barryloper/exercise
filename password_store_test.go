package main

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
func TestUpdateAverageHashTime(t *testing.T) {
	// Tests that updating the average asynchronously results in the expected average

	t.Parallel()

	stats := &HashStats{}
	var times [100]time.Duration //an array of integers representing seconds
	var sumtimes time.Duration
	for i := range times {
		times[i] = time.Duration(rand.Int63n(24)) * time.Hour
		sumtimes += times[i]

	}
	expectedAverage := sumtimes / time.Duration(len(times)) // expected average time

	for _, t := range times {
		// updates the average time with a bunch of goroutines
		go func(t time.Duration) {
			stats.hashInFlight()
			stats.hashComplete(time.Duration(t))
		}(t)
	}
	stats.hashesInFlight.Wait()

	if stats.averageHashTime-expectedAverage > time.Duration(1)*time.Millisecond {
		t.Log("Wanted average to be ", expectedAverage, " got ", stats.averageHashTime)
		t.Fail()
	}

}

func TestAddHash(t *testing.T) {
	t.Parallel()
	hashTable := NewPasswordStore()

	for i := 0; i < numHashesToTest; i++ {

		go func() {
			password := RandomPassword(minPasswordLength, maxPasswordLength)
			_, err := hashTable.SavePassword(password)
			if err != nil {
				t.Error("Error beginning the hash add")
				t.Fail()
			}
			return
		}()

	}
	hashTable.stats.hashesInFlight.Wait()

	count, duration := hashTable.GetStats()
	t.Log("Added", count, "hashes. Avg time:", duration)
}
