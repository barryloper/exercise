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
	lock sync.RWMutex
}

type HashStats struct {
	averageHashTime time.Duration
	passwordCount   int // helps with calculating average duration since len(PasswordStore.passwords) might include empty hashes
	hashesInFlight  sync.WaitGroup
	lock            sync.RWMutex
}

// PasswordStore stores an array of hashes along with a mutex for concurrent access of
// that array
type PasswordStore struct {
	passwords []*Password
	lock      sync.RWMutex
	stats     *HashStats
}

// NewPasswordStore constructor of PasswordStore for convention
func NewPasswordStore() *PasswordStore {
	return &PasswordStore{stats: &HashStats{}}
}

// SavePassword salts, hashes, and saves a password to the password store
// It returns the ID of that hash for later retrieval.
// It returns a channel on which you can wait to receive the hash once calculated
func (store *PasswordStore) SavePassword(password []byte) (int, error) {
	var err error
	store.lock.Lock() // need to lock so we get a valid hash number
	defer store.lock.Unlock()

	store.stats.hashInFlight()

	hashID := len(store.passwords) // the new hash will be this element in the hash table
	newPassword := Password{salt: make([]byte, hashsize)}
	_, err = rand.Read(newPassword.salt)
	if err != nil {
		return hashID, err
	}
	store.passwords = append(store.passwords, &newPassword) // zero'd hash for now

	// compute the hash in the background
	go func(p *Password) {
		time.Sleep(5 * time.Second)
		start := time.Now()
		saltedPass := append(password, p.salt...)
		computed := hashfunction(saltedPass) // todo: avoid copying this hash around?
		elapsed := time.Since(start)
		store.stats.hashComplete(elapsed)

		p.updateHash(computed)

	}(&newPassword)

	return hashID, err
}

// GetHash takes a userID and returns the associated password hash
func (store *PasswordStore) GetHash(number int) ([hashsize]byte, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()
	user, err := store.checkoutUser(number)
	if err != nil {
		return [hashsize]byte{}, err
	}
	defer store.checkinUser(number)

	return user.hash, nil
}

// CheckPassword compares the salted hash of the supplied password to the one from the password store
func (store *PasswordStore) CheckPassword(userID int, password []byte) bool {
	store.lock.RLock()
	user, err := store.checkoutUser(userID)
	store.lock.RUnlock()
	if err != nil {
		return false
	}
	// don't need to check in user if err != nil
	defer store.checkinUser(userID)

	providedHash := hashfunction(append(password, user.salt...))
	return providedHash == user.hash
}

// GetStats returns the number of passwords and average password hashing time from the password store
func (store *PasswordStore) GetStats() (int, time.Duration) {
	store.stats.lock.RLock()
	defer store.stats.lock.RUnlock()
	return store.stats.passwordCount, store.stats.averageHashTime
}

// Private methods ----------------------------

// Returns a user locked for read. Unlock the user once you are done
func (store *PasswordStore) checkoutUser(userID int) (*Password, error) {
	store.lock.RLock() // needed to consistently read len(h.passwords)
	defer store.lock.RUnlock()

	if userID < len(store.passwords) {
		// locking here so the caller receives a locked entity
		store.passwords[userID].lock.Lock()
		return store.passwords[userID], nil
	}
	return &Password{}, errors.New("Invalid user ID")
}

func (store *PasswordStore) checkinUser(userID int) error {
	store.lock.RLock() // needed to consistently read len(h.passwords)
	defer store.lock.RUnlock()

	if userID < len(store.passwords) {
		// locking here so the caller receives a user locked for read
		store.passwords[userID].lock.Unlock()
		return nil
	}
	return errors.New("Invalid user ID")
}

func (stats *HashStats) hashInFlight() {
	stats.lock.Lock()
	defer stats.lock.Unlock()
	stats.hashesInFlight.Add(1)
}

func (stats *HashStats) hashComplete(hashDuration time.Duration) {
	stats.lock.Lock()
	defer stats.lock.Unlock()

	stats.passwordCount++
	stats.averageHashTime = stats.averageHashTime + (hashDuration-stats.averageHashTime)/time.Duration(stats.passwordCount)
	stats.hashesInFlight.Done()
	return

}

func (p *Password) updateHash(hash [hashsize]byte) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.hash = hash
}
