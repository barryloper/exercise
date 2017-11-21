package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestPasswordStore_updateAverageHashTime(t *testing.T) {
	// Tests that updating the average asynchronously results in the expected average

	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	store := PasswordStore{}
	var times [100]time.Duration //an array of integers representing seconds
	var sumtimes time.Duration
	for i := range times {
		times[i] = time.Duration(r.Int63n(24)) * time.Hour
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
