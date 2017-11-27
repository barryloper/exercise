package main

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"sync"
	"time"
)

// makes changing hash function a little easier
const hashsize int = sha512.Size

var hashfunction = sha512.Sum512

// Password is an element of the array PasswordStore.passwords
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
	passwordsLock   sync.RWMutex
	statsLock       sync.RWMutex
	hashesComputing sync.WaitGroup
}

// NewPasswordStore constructor of PasswordStore for convention
func NewPasswordStore() *PasswordStore {
	store := &PasswordStore{}
	return store
}

// Sync returns when all pending have been computed
// be aware, more hashes may be added immediately after this returns
func (h *PasswordStore) Sync() {
	h.hashesComputing.Wait()
	return
}

// SavePassword salts, hashes, and saves a password to the password store
// It returns the ID of that hash for later retrieval.
// It returns a channel on which you can wait to receive the hash once calculated
func (h *PasswordStore) SavePassword(password []byte) (int, <-chan [hashsize]byte, error) {
	return h.addHash(password)
}

// CheckPassword compares the salted hash of the supplied password to the one from the password store
func (h *PasswordStore) CheckPassword(user int, password []byte) bool {
	h.passwordsLock.RLock()
	defer h.passwordsLock.RUnlock()

	//fmt.Println("hash address ", &h)
	salt, err := h.getSalt(user)
	if err != nil {
		return false
	}
	theirPass := hashfunction(append(password, salt...))
	dbPass, err := h.getHash(user)
	if err != nil {
		return false
	}
	//fmt.Printf("expected %+v\n", dbPass)
	//fmt.Printf("got %+v\n", theirPass)
	return dbPass == theirPass
}

// GetStats returns the number of passwords and average password hashing time from the password store
func (h *PasswordStore) GetStats() (int64, time.Duration) {
	h.statsLock.RLock()
	defer h.statsLock.RUnlock()
	return h.passwordCount, h.averageHashTime
}

// Private methods ----------------------------

func (h *PasswordStore) updateAverageHashTime(hashDuration time.Duration) {
	h.statsLock.Lock()
	defer h.statsLock.Unlock()

	newCount := h.passwordCount + 1
	oldAverage := h.averageHashTime
	h.averageHashTime = oldAverage + (hashDuration-oldAverage)/time.Duration(newCount)
	h.passwordCount = newCount
	return

}

func (h *PasswordStore) addHash(password []byte) (int, <-chan [hashsize]byte, error) {
	var err error
	h.passwordsLock.Lock() // need to lock so we get a valid hash number
	defer h.passwordsLock.Unlock()

	hashID := len(h.passwords) // the new hash will be this element in the hash table
	newSalt := make([]byte, hashsize)
	_, err = rand.Read(newSalt)
	if err != nil {
		return hashID, nil, err
	}
	h.passwords = append(h.passwords, Password{salt: newSalt}) // zero'd hash for now

	hashValueChannel := make(chan [hashsize]byte)

	// compute the hash in the background
	// increment a waitgroup so the program doesn't terminate before finishing the hash
	// remember that you can't Add() to a waitgroup that has been Wait()ed
	h.hashesComputing.Add(1)
	go func() {
		defer h.hashesComputing.Done()
		time.Sleep(5 * time.Second)
		start := time.Now()
		saltedPass := append(password, newSalt...)
		computed := hashfunction(saltedPass)
		elapsed := time.Since(start)
		h.updateAverageHashTime(elapsed)
		//fmt.Println("adding hash to ", h)
		hashEntry := &h.passwords[hashID]
		hashEntry.hash = computed

		hashValueChannel <- hashEntry.hash // callers may optionally wait on this channel and get the hash when it is done
	}()

	return hashID, hashValueChannel, err
}

func (h *PasswordStore) getHash(number int) ([hashsize]byte, error) {
	h.passwordsLock.RLock()
	defer h.passwordsLock.RUnlock()

	if number >= len(h.passwords) {
		return [hashsize]byte{}, errors.New("User not found")
	}
	return h.passwords[number].hash, nil
}

func (h *PasswordStore) getSalt(number int) ([]byte, error) {
	//fmt.Println("getting salt from ", &h)
	h.passwordsLock.RLock()
	defer h.passwordsLock.RUnlock()

	if number >= len(h.passwords) {
		return []byte{}, errors.New("User not found")
	}

	return h.passwords[number].salt, nil
}
