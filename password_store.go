package main

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// makes changing hash function a little easier
const hashsize int = sha512.Size
const shutdownTimeout time.Duration = 30 * time.Second

var hashfunction = sha512.Sum512

// Password is an element of the array PasswordStore.passwords
type Password struct {
	hash [hashsize]byte
	salt []byte
}

// PasswordStore stores an array of hashes, stats and synchronization aparatus
type PasswordStore struct {
	passwords             map[int]*Password
	nextUserID            chan int
	nextValidUserID       int
	updatedPasswords      chan newHash
	userQuery             chan int
	userResponse          chan Password
	statsQuery            chan bool
	statsResponse         chan StatsBody
	hashesInFlight        sync.WaitGroup
	signals               chan os.Signal
	averageHashTime       time.Duration
	completePasswordCount uint
}

// StatsBody defines the structure of stats data for both JSON and password store manager queries
type StatsBody struct {
	Total   uint          `json:"total"`
	Average time.Duration `json:"average"`
}

// newHash used for message passing in PasswordStore.startStoreManager
type newHash struct {
	ID          int
	password    *Password
	computeTime time.Duration
}

// NewPasswordStore creates a PasswordStore and starts the manager loop
func NewPasswordStore() *PasswordStore {
	store := &PasswordStore{
		nextValidUserID:  1,
		passwords:        make(map[int]*Password),
		nextUserID:       make(chan int),
		updatedPasswords: make(chan newHash),
		userQuery:        make(chan int),
		userResponse:     make(chan Password),
		statsQuery:       make(chan bool),
		statsResponse:    make(chan StatsBody),
		signals:          make(chan os.Signal, 1),
	}
	signal.Notify(store.signals, os.Interrupt, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	go store.startStoreManager()
	return store
}

// SavePassword salts, hashes, and saves a password to the password store
// It returns the ID of that hash for later retrieval.
// It returns a channel on which you can wait to receive the hash once calculated
func (store *PasswordStore) SavePassword(password []byte) (int, <-chan [hashsize]byte) {
	newID := <-store.nextUserID
	response := make(chan [hashsize]byte)

	// compute the hash in the background
	ComputeHash := func(newID int, password []byte, response chan [hashsize]byte) {
		time.Sleep(5 * time.Second) // part of the spec.

		start := time.Now()

		p := Password{}
		p.salt = make([]byte, hashsize)
		_, err := rand.Read(p.salt)
		if err != nil {
			log.Fatal("Error generating random salt.", err)
		}
		p.hash = hashfunction(append(password, p.salt...))

		store.updatedPasswords <- newHash{newID, &p, time.Since(start)}
		response <- p.hash
	}
	go ComputeHash(newID, password, response)

	return newID, response
}

// GetHash takes a userID and returns the associated password hash
func (store *PasswordStore) GetHash(userID int, timeout time.Duration) ([hashsize]byte, error) {
	store.userQuery <- userID
	select {
	case user := <-store.userResponse:
		return user.hash, nil
	case <-time.After(timeout):
		return [hashsize]byte{}, errors.New(fmt.Sprint("User ", userID, " not found."))
	}
}

// CheckPassword compares the salted hash of the supplied password to the one from the password store
func (store *PasswordStore) CheckPassword(userID int, password []byte) bool {
	store.userQuery <- userID
	user := <-store.userResponse
	hashOfPassword := hashfunction(append(password, user.salt...))
	return hashOfPassword == user.hash
}

// GetStats returns the number of passwords and average password hashing time from the password store
func (store *PasswordStore) GetStats() StatsBody {
	store.statsQuery <- true
	stats := <-store.statsResponse
	return stats
}

func (store *PasswordStore) startStoreManager() {
	// no other goroutines should modify the password store

	for {
		select {
		case store.nextUserID <- store.nextValidUserID: // this channel generates user IDs for SavePassword()
			store.hashesInFlight.Add(1) // useful to wait for hashes that are being computed
			store.nextValidUserID++
		case delUserID := <-store.nextUserID: //receiving an ID here is a signal to delete
			delete(store.passwords, delUserID)
		case <-store.statsQuery:
			// syncs read access to stats variables
			store.statsResponse <- StatsBody{Total: store.completePasswordCount, Average: store.averageHashTime}
		case updateRequest := <-store.updatedPasswords:
			// new passwords or updated passwords are sent throgh this channel
			store.passwords[updateRequest.ID] = updateRequest.password
			store.completePasswordCount++
			store.averageHashTime = (store.averageHashTime +
				(updateRequest.computeTime-store.averageHashTime)/
					time.Duration(store.completePasswordCount))
			store.hashesInFlight.Done()
		case userID := <-store.userQuery:
			// avoids race to read/write users
			response := store.passwords[userID]
			if response != nil { //avoid dereferencing nil pointer
				store.userResponse <- *response
			}
			// don't return anything on userResponse for invalid user
			// callers should timeout
			// useful for tarpitting on the http side
		case sig := <-store.signals:
			log.Println("Caught", sig)
			store.gracefulShutdown()
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

}

func (store *PasswordStore) gracefulShutdown() {
	done := make(chan bool)
	go func(chan bool) {
		log.Println("Attempting to shut down gracefully.")
		store.hashesInFlight.Wait()
		done <- true
	}(done)
	select {
	case <-done:
		log.Println("Shut down gracefully.")
		os.Exit(0)
	case <-time.After(shutdownTimeout):
		log.Fatal("Timeout waiting for inflight hashes during shutdown.")
	}
}
