package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
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

func TestHashEndpoint(t *testing.T) {
	hashUrl := fmt.Sprint(testServer.URL, "/hash/")
	body, marshalErr := json.Marshal(passwordBody{"foo bar"})
	if marshalErr != nil {
		t.Fatal("Failed to encode password")
	}
	response, httpErr := testServer.Client().Post(hashUrl, "text/json", bytes.NewBuffer(body))
	//todo close response.body after reading from it?
	if httpErr != nil {
		t.Fatal("Request failed", httpErr.Error())
	}
	responseBody, _ := ioutil.ReadAll(response.Body)
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
		response, _ := ioutil.ReadAll(getResponse.Body)
		t.Fatal("get user id failed", string(response))
	}
	getResponseBody, _ := ioutil.ReadAll(getResponse.Body)
	unmartialErr2 := json.Unmarshal(getResponseBody, userInfo)
	if unmartialErr2 != nil {
		t.Fatal("Invalid response to get", string(getResponseBody))
	}

	if userInfo.PasswordHash != "" {

		t.Log("Hash for user", userInfo.UserID, "is", userInfo.PasswordHash)
	}
}

//The software should be able to process multiple connections simultaneously
func TestSimultaneousRequest(t *testing.T) {}

//The software should support a graceful shutdown request.
//it should allow any remaining password hashing to complete,
//reject any new requests, and shutdown.
func TestGracefulShutdown(t *testing.T) {}

//No additional password requests should be allowed when shutdown is pending.
func TestShuttingDownPasswordRequest(t *testing.T) {}
