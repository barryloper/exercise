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
func TestUpdateAverageHashTime(t *testing.T) {
	// Tests that updating the average asynchronously results in the expected average

	// t.Parallel()

	// var times [100]time.Duration //an array of integers representing seconds
	// var sumtimes time.Duration
	// for i := range times {
	// 	times[i] = time.Duration(rand.Int63n(24)) * time.Hour
	// 	sumtimes += times[i]

	// }
	// expectedAverage := sumtimes / time.Duration(len(times)) // expected average time

	// for _, t := range times {
	// 	// updates the average time with a bunch of goroutines
	// 	stats.hashInFlight()
	// 	go func(t time.Duration) {
	// 		stats.hashComplete(time.Duration(t))
	// 	}(t)
	// }
	// stats.hashesInFlight.Wait()

	// if stats.averageHashTime-expectedAverage > time.Duration(1)*time.Millisecond {
	// 	t.Log("Wanted average to be ", expectedAverage, " got ", stats.averageHashTime)
	// 	t.Fail()
	// }

}

func TestAddHash(t *testing.T) {
	t.Parallel()
	hashTable := NewPasswordStore()
	var wg sync.WaitGroup
	wg.Add(numHashesToTest)
	for i := 0; i < numHashesToTest; i++ {

		go func() {
			defer wg.Done()
			password := RandomPassword(minPasswordLength, maxPasswordLength)
			hashTable.SavePassword(password)
			return
		}()

	}
	wg.Wait()
	hashTable.hashesInFlight.Wait()

	stats := hashTable.GetStats()
	t.Log("Added", stats.count, "hashes. Avg time:", stats.averageHashTime)
}
