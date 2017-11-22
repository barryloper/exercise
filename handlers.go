package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// global password store is blank on startup
var passwordStore *PasswordStore = &PasswordStore{}

// response json body of a GET to the password handler
type credential struct {
	userID       int    `json:"userId"`
	passwordHash string `json:"passwordHash"`
}

// expected json body of a POST to the password handler
type password struct {
	password string `json:"password"`
}

func getCredential(id int) *credential {
	hash := passwordStore.getHash(id)
	encodedHash := base64.StdEncoding.EncodeToString(hash[:])
	return &credential{id, encodedHash}
}

func storePassword(password *password) (int, error) {
	userID, _, err := passwordStore.addHash([]byte(password.password))
	return userID, err
}

//DocsHandler returns some documentation to help users discover the other endpoints
func DocsHandler(w http.ResponseWriter, r *http.Request) {}

/*HashHandler
  A POST to /hash should accept a password; it should return a job identifier immediate;
  it should then wait 5 seconds and compute the password hash.
  The hashing algorithm should be SHA512.

  A GET to /hash should accept a job identifier;
  it should return the base64 encoded password hash for the corresponding POST request.
*/
func HashHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json")

	switch r.Method {
	case http.MethodGet:
		userID := r.URL.Path[len("/hash/"):]
		if len(userID) > 0 {
			id, err := strconv.Atoi(userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "Invalid url")
				return
			}
			credentials := getCredential(id)
			if credentials != nil {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(credentials)
			}
		}

	case http.MethodPost:
		userPassword := password{}
		var jobID int
		var err error
		err = json.NewDecoder(r.Body).Decode(&userPassword)
		jobID, err = storePassword(&userPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		credentials := credential{userID: jobID}
		json.NewEncoder(w).Encode(credentials)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	// POST
	// return a job identifier (to find the password hash later)
	// wait 5 seconds
	// compute sha512 password hash and store it associated with the job id
	// store number of milliseconds it took the hash to complete

	// GET
	// return base64 encoded hash for job id
}

/*StatsHandler
A GET to /stats should accept no data;
it should return a JSON data structure for the total hash requests since server start
and the average time of a hash request in milliseconds, not including the 500ms wait.
*/
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	// GET
	// returns stats in the following example format
	/*
	  {
	    "total": 24,
	    "average": 123
	  }
	*/
	io.WriteString(w, "Hello from Stats")
}
