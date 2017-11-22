package main

import (
	"crypto/rand"
	"crypto/sha512"
	"sync"
	"time"
)

// makes changing hash function a little easier
const hashsize int = sha512.Size

var hashfunction = sha512.Sum512

//if the hashes live in a slice, or a map, need a function that will
// block
// add an item to the slice
// spawn a goroutine that modifies that item
// return the id of the item
type Password struct {
	hash [hashsize]byte
	salt []byte
}

// PasswordStore stores an array of hashes along with a mutex for concurrent access of
// that array
type PasswordStore struct {
	passwords       []Password
	averageHashTime time.Duration
	passwordCount   int64 // helps with calculating average duration since len(passwords) might include empty hashes
	mutex           sync.Mutex
}

func (h *PasswordStore) updateAverageHashTime(hashDuration time.Duration) {
	// keep concurrency in mind
	h.mutex.Lock()
	defer h.mutex.Unlock()

	newCount := h.passwordCount + 1
	oldAverage := &h.averageHashTime
	h.averageHashTime = *oldAverage + (hashDuration-*oldAverage)/time.Duration(newCount)
	h.passwordCount = newCount

	return

}

func (h *PasswordStore) addHash(password []byte) (int, <-chan [hashsize]byte, error) {
	var err error = nil
	h.mutex.Lock()             // need to lock so we get a valid hash number
	hashID := len(h.passwords) // the new hash will be this element in the hash table
	newSalt := make([]byte, hashsize)
	_, err = rand.Read(newSalt)
	if err != nil {
		return hashID, nil, err
	}
	h.passwords = append(h.passwords, Password{salt: newSalt}) // zero'd hash for now
	h.mutex.Unlock()

	hashValueChannel := make(chan [hashsize]byte)

	go func() {
		start := time.Now()
		saltedPass := append(password, newSalt...)
		computed := hashfunction(saltedPass)
		time.Sleep(5 * time.Second)
		elapsed := time.Since(start)
		h.updateAverageHashTime(elapsed)
		//fmt.Println("adding hash to ", h)
		hashEntry := &h.passwords[hashID]
		hashEntry.hash = computed

		hashValueChannel <- hashEntry.hash // callers may optionally wait on this channel and get the hash
	}()

	return hashID, hashValueChannel, err
}

func (h *PasswordStore) getHash(number int) [hashsize]byte {
	return h.passwords[number].hash
}

func (h *PasswordStore) getSalt(number int) []byte {
	//fmt.Println("getting salt from ", &h)
	return h.passwords[number].salt
}

func (h *PasswordStore) checkPassword(user int, password []byte) bool {
	//fmt.Println("hash address ", &h)
	salt := h.getSalt(user)
	theirPass := hashfunction(append(password, salt...))
	dbPass := h.getHash(user)
	//fmt.Printf("expected %+v\n", dbPass)
	//fmt.Printf("got %+v\n", theirPass)
	return dbPass == theirPass
}

func (h *PasswordStore) getDuration(number int) {
	return //h.table[number].computeTime
}
