package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

var testServer *httptest.Server

func TestMain(m *testing.M) {
	db := NewPasswordStore()
	muxer := MakeMuxer(db)
	testServer = httptest.NewServer(muxer)
	defer testServer.Close()
	os.Exit(m.Run())
}

func TestHashRequests(t *testing.T) {
	t.Parallel()

	hashUrl := fmt.Sprint(testServer.URL, "/hash/")

	wg := sync.WaitGroup{}
	wg.Add(numHashesToTest)

	tryPassword := func(password string) {

		body, marshalErr := json.Marshal(passwordBody{password})
		if marshalErr != nil {
			t.Fatal("Failed to encode password")
		}
		response, httpErr := testServer.Client().Post(hashUrl, "text/json", bytes.NewBuffer(body))
		if httpErr != nil {
			t.Fatal("Request failed", httpErr.Error())
		}
		responseBody, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		userInfo := &credentialBody{}
		unmartialErr := json.Unmarshal(responseBody, userInfo)
		if unmartialErr != nil {
			t.Fatal("Invalid response to post", string(responseBody))
		}
		if userInfo.PasswordHash != "" {
			t.Fatal("Initial post returned non-nil hash", userInfo.PasswordHash)
		}
		newUserId := userInfo.UserID

		time.Sleep(8 * time.Second)

		// check the pasword was created
		getResponse, getErr := testServer.Client().Get(fmt.Sprint(hashUrl, "/", newUserId))
		if getErr != nil {
			t.Fatal("get user id failed", getErr.Error())
		}
		getResponseBody, _ := ioutil.ReadAll(getResponse.Body)
		getResponse.Body.Close()
		unmartialErr2 := json.Unmarshal(getResponseBody, userInfo)
		if unmartialErr2 != nil {
			t.Fatal("Invalid response to get", string(getResponseBody))
		}

		t.Log("Hash response", fmt.Sprintf("%s", getResponseBody))

		wg.Done()
	}

	for i := 0; i < numHashesToTest; i++ {
		password := RandomPassword(minPasswordLength, maxPasswordLength)
		stringPassword := fmt.Sprintf("%s", password)
		go tryPassword(stringPassword)

	}
	wg.Wait()
}

//The software should support a graceful shutdown request.
//it should allow any remaining password hashing to complete,
//reject any new requests, and shutdown.
func TestGracefulShutdown(t *testing.T) {}

//No additional password requests should be allowed when shutdown is pending.
func TestShuttingDownPasswordRequest(t *testing.T) {}
