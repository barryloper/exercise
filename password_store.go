package main

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"log"
	"sync"
	"time"
)

// makes changing hash function a little easier
const hashsize int = sha512.Size
const hashBufferSize int = 10

var hashfunction = sha512.Sum512

// Password is an element of the array PasswordStore.passwords
type Password struct {
	hash [hashsize]byte
	salt []byte
}

// PasswordStore stores an array of hashes along with a mutex for concurrent access of
// that array
type PasswordStore struct {
	passwords             []*Password
	newUsers              chan blankPassword
	updatedPasswords      chan newHash
	userQuery             chan int
	userResponse          chan Password
	errorResponse         chan bool
	statsQuery            chan bool
	statsResponse         chan statsMessage
	hashesInFlight        sync.WaitGroup
	averageHashTime       time.Duration
	completePasswordCount uint
}

// passwordTemplate is the structure useful for creating new passwords
type blankPassword struct {
	ID   int
	salt []byte
}

type newHash struct {
	ID          int
	hash        [hashsize]byte
	computeTime time.Duration
}

type statsMessage struct {
	count           uint
	averageHashTime time.Duration
}

// NewPasswordStore constructor of PasswordStore for convention
func NewPasswordStore() *PasswordStore {
	store := &PasswordStore{
		newUsers:         make(chan blankPassword),
		updatedPasswords: make(chan newHash, hashBufferSize),
		userQuery:        make(chan int),
		userResponse:     make(chan Password),
		statsQuery:       make(chan bool),
		statsResponse:    make(chan statsMessage),
	}
	go store.startStoreManager()
	return store
}

// SavePassword salts, hashes, and saves a password to the password store
// It returns the ID of that hash for later retrieval.
// It returns a channel on which you can wait to receive the hash once calculated
func (store *PasswordStore) SavePassword(password []byte) int {
	newTemplate := <-store.newUsers

	// compute the hash in the background
	ComputeHash := func(template blankPassword, password []byte) {
		time.Sleep(5 * time.Second) // part of the spec.

		start := time.Now()
		hash := hashfunction(append(password, template.salt...))
		duration := time.Since(start)

		store.updatedPasswords <- newHash{template.ID, hash, duration}

	}
	go ComputeHash(newTemplate, password)

	return newTemplate.ID
}

// GetHash takes a userID and returns the associated password hash
func (store *PasswordStore) GetHash(number int) ([hashsize]byte, error) {
	store.userQuery <- number
	select {
	case user := <-store.userResponse:
		return user.hash, nil
	case <-store.errorResponse:
		return [hashsize]byte{}, errors.New("Invalid UserID")
	}
}

// CheckPassword compares the salted hash of the supplied password to the one from the password store
func (store *PasswordStore) CheckPassword(userID int, password []byte) bool {
	store.userQuery <- userID
	select {
	case user := <-store.userResponse:
		providedHash := hashfunction(append(password, user.salt...))
		return providedHash == user.hash
	case <-store.errorResponse:
		return false
	}
}

// GetStats returns the number of passwords and average password hashing time from the password store
func (store *PasswordStore) GetStats() statsMessage {
	store.statsQuery <- true
	stats := <-store.statsResponse
	return stats
}

func newBlankPassword(ID int) blankPassword {
	var p blankPassword
	p.ID = ID
	p.salt = make([]byte, hashsize)
	_, err := rand.Read(p.salt)
	if err != nil {
		log.Fatal("Error generating random salt.", err)
	}
	return p
}

func (store *PasswordStore) startStoreManager() {
	// no other goroutines should modify the password store
	bp := newBlankPassword(len(store.passwords)) // first blank password for the newUsers channel
	for {
		select {
		case store.newUsers <- bp:
			store.passwords = append(store.passwords, &Password{salt: bp.salt})
			store.hashesInFlight.Add(1)
			bp = newBlankPassword(len(store.passwords))
		case <-store.statsQuery:
			store.statsResponse <- statsMessage{count: store.completePasswordCount, averageHashTime: store.averageHashTime}
		case updateRequest := <-store.updatedPasswords:
			store.passwords[updateRequest.ID].hash = updateRequest.hash
			store.completePasswordCount++
			store.averageHashTime = store.averageHashTime + (updateRequest.computeTime-store.averageHashTime)/time.Duration(store.completePasswordCount)
			store.hashesInFlight.Done()
		case userID := <-store.userQuery:
			if userID < len(store.passwords) {
				store.userResponse <- *store.passwords[userID]
			} else {
				store.errorResponse <- true
			}
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

}
