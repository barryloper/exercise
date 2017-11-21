package main

import (
	"crypto/rand"
	"crypto/sha512"
	"log"
	"sync"
	"time"
)

func hashSaltedPass(saltedPass []byte) [sha512.Size]byte {
	// abstract what hash we will be using
	return sha512.Sum512(saltedPass)
}

//if the hashes live in a slice, or a map, need a function that will
// block
// add an item to the slice
// spawn a goroutine that modifies that item
// return the id of the item
type hash struct {
	hash        [sha512.Size]byte
	salt        []byte
	computeTime time.Duration
}

// HashTable stores an array of hashes along with a mutex for concurrent access of
// that array
type HashTable struct {
	table []hash
	mutex sync.Mutex
}

func (h *HashTable) addHash(password []byte) int {
	h.mutex.Lock()         // need to lock so we get a valid hash number
	hashID := len(h.table) // the new hash will be this element in the hash table
	newSalt := make([]byte, sha512.Size)
	_, err := rand.Read(newSalt)
	if err != nil {
		log.Fatal("Couldn't generate salt.")
	}
	h.table = append(h.table, hash{salt: newSalt}) // zero'd hash for now
	h.mutex.Unlock()

	//todo salt

	go func() {
		start := time.Now()
		saltedPass := append(password, newSalt...)
		computed := hashSaltedPass(saltedPass)
		time.Sleep(5 * time.Second)
		elapsed := time.Since(start)
		//fmt.Println("adding hash to ", h)
		hashEntry := &h.table[hashID]
		hashEntry.hash = computed
		hashEntry.computeTime = elapsed
	}()
	return hashID
}

func (h *HashTable) getHash(number int) [sha512.Size]byte {
	return h.table[number].hash
}

func (h *HashTable) getSalt(number int) []byte {
	//fmt.Println("getting salt from ", &h)
	return h.table[number].salt
}

func (h *HashTable) checkPassword(user int, password []byte) bool {
	//fmt.Println("hash address ", &h)
	salt := h.getSalt(user)
	theirPass := hashSaltedPass(append(password, salt...))
	dbPass := h.getHash(user)
	//fmt.Printf("expected %+v\n", dbPass)
	//fmt.Printf("got %+v\n", theirPass)
	return dbPass == theirPass
}

func (h *HashTable) getDuration(number int) time.Duration {
	return h.table[number].computeTime
}
