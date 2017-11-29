package main

import (
	"math/rand"
)

const numHashesToTest int = 100
const minPasswordLength int = 6
const maxPasswordLength int = 32
const defaultMinPasswordLength int = 6
const defaultMaxPasswordLength int = 32

// RandomPassword generates a byte slice of random bytes
// slice length is between 1 and maxPasswordLengthBytes
// Not cryptographically secure. Call math/rand.Seed(seed int64) in the calling program
func RandomPassword(minPasswordLengthBytes, maxPasswordLengthBytes int) []byte {
	passwordLength := minPasswordLengthBytes + rand.Intn(maxPasswordLengthBytes-minPasswordLengthBytes)
	password := make([]byte, passwordLength)
	rand.Read(password)
	return password
}
