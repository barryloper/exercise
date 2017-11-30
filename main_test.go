package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const numHashesToTest int = 20

// numHashesToTest is kind of low because t.Parallel() seems to go on forever with high numbers
// using regular goroutines and waitgroups goes fast as expected, though, and thousands work fine
// todo: maybe testing.B would work better to analyze how many requests affect the system
const minPasswordLength int = 6
const maxPasswordLength int = 32
const defaultMinPasswordLength int = 6
const defaultMaxPasswordLength int = 32

// RandomPassword generates a byte slice of random bytes
// slice length is between 1 and maxPasswordLengthBytes
// Not limited to displayable characters; it should test that we are able to hash arbitrary passwords of reasonable length
// Not cryptographically secure. Call math/rand.Seed(seed int64) in the calling program
func RandomPassword(minPasswordLengthBytes, maxPasswordLengthBytes int) []byte {
	passwordLength := minPasswordLengthBytes + rand.Intn(maxPasswordLengthBytes-minPasswordLengthBytes)
	password := make([]byte, passwordLength)
	rand.Read(password)
	return password
}

var testServer *httptest.Server

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	db := NewPasswordStore()
	muxer := MakeMuxer(db)
	testServer = httptest.NewServer(muxer)
	defer testServer.Close()
	os.Exit(m.Run())
}

func TestHashRequests(t *testing.T) {

	hashURL := fmt.Sprint(testServer.URL, "/hash/")

	tryPassword := func(password string, t *testing.T) {
		t.Parallel()

		body, marshalErr := json.Marshal(password)
		if marshalErr != nil {
			t.Fatal("Failed to encode password")
		}
		response, httpErr := testServer.Client().Post(hashURL, "text/json", bytes.NewBuffer(body))
		if httpErr != nil {
			t.Fatal("Request failed", httpErr.Error())
		}
		responseBody, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		var userInfo int
		unmartialErr := json.Unmarshal(responseBody, &userInfo)
		if unmartialErr != nil {
			err := unmartialErr.Error()
			t.Fatal("Invalid response to post", err)
		}

		time.Sleep(8 * time.Second)

		// check the pasword was created
		// Only sleeping 8 seconds ensures this will fail if the server is not processing requests
		// in parallel, since a hash will take between 5 and 6 seconds to process.
		getResponse, getErr := testServer.Client().Get(fmt.Sprint(hashURL, "/", userInfo))
		if getErr != nil {
			t.Fatal("get user id failed", getErr.Error())
		}
		getResponseBody, _ := ioutil.ReadAll(getResponse.Body)
		getResponse.Body.Close()

		var userHash string
		unmartialErr2 := json.Unmarshal(getResponseBody, &userHash)
		if unmartialErr2 != nil {
			err := unmartialErr2.Error()
			t.Fatal("Invalid response to get", err)
		}
		//t.Log("Hash response", fmt.Sprintf("%s", getResponseBody))
	}

	for i := 0; i < numHashesToTest; i++ {
		password := RandomPassword(minPasswordLength, maxPasswordLength)
		stringPassword := fmt.Sprintf("%s", password)
		t.Run(stringPassword, func(t *testing.T) { tryPassword(stringPassword, t) })

	}
}

//The software should support a graceful shutdown request.
//it should allow any remaining password hashing to complete,
//reject any new requests, and shutdown.
func TestGracefulShutdown(t *testing.T) {}

//No additional password requests should be allowed when shutdown is pending.
func TestShuttingDownPasswordRequest(t *testing.T) {}
