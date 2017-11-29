package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MakeMuxer sets up the routes and methods and returns a configured multiplexer
// requires a pointer to a PasswordStore instance
func MakeMuxer(db *PasswordStore) *http.ServeMux {
	mux := http.NewServeMux()

	// todo: this still seems a bit clunky. how can this be done better?
	hashMethodHandlers := make(methodMap, 2)
	hashMethodHandlers[http.MethodGet] = getHash
	hashMethodHandlers[http.MethodPost] = addHash

	statsMethodHandlers := make(methodMap, 1)
	statsMethodHandlers[http.MethodGet] = getStats

	hashHandler := restEndpoint{
		route:   "/hash/",
		methods: hashMethodHandlers,
		db:      db,
	}

	statsHandler := restEndpoint{
		route:   "/stats",
		methods: statsMethodHandlers,
		db:      db,
	}

	mux.Handle(hashHandler.route, hashHandler)
	mux.Handle(statsHandler.route, statsHandler)
	return mux
}

// body of a /hash response
type credentialBody struct {
	UserID       int    `json:"userId"`
	PasswordHash string `json:"passwordHash,omitempty"` // []byte is marshalled into a base64 encoded string
}

// expected body of a POST to the /hash handler
type passwordBody struct {
	Password string `json:"password"`
}

// body of a /stats response
type statsBody struct { // todo: don't know if json can encode these types
	Total   uint          `json:"total"`
	Average time.Duration `json:"average"`
}

// A methodHandler function should return a value encodable via json.Encode, along with an http.Status* constant
type methodHandler func([]string, *PasswordStore, *http.Request) (interface{}, int)

// strings in methodMap should be one of the http method constants defined in the http package
type methodMap map[string]methodHandler

// restEndpoint implements http.Handler interface
// it is configured with a MethodMap, route and given access to a database
type restEndpoint struct {
	route   string
	methods methodMap
	db      *PasswordStore
}

func (endpoint restEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	pathArgs := strings.Split(r.URL.Path[len(endpoint.route):], "/")

	methodFn := endpoint.methods[r.Method]

	if methodFn == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	value, httpErr := endpoint.methods[r.Method](pathArgs, endpoint.db, r)

	w.WriteHeader(httpErr)
	json.NewEncoder(w).Encode(value)

}

/*HashHandler
  A POST to /hash should accept a password; it should return a job identifier immediate;
  it should then wait 5 seconds and compute the password hash.
  The hashing algorithm should be SHA512.

  A GET to /hash should accept a job identifier;
  it should return the base64 encoded password hash for the corresponding POST request.
*/
func getHash(pathArgs []string, db *PasswordStore, r *http.Request) (interface{}, int) {
	// GET
	// return base64 encoded hash for job id
	userID := pathArgs[0]
	if len(userID) > 0 {
		id, err := strconv.Atoi(userID)
		if err != nil {
			return "User ID must be an integer", http.StatusInternalServerError
		}

		hash, err := db.GetHash(id)
		if err != nil {
			return "User not found", http.StatusNotFound
		}
		encodedHash := base64.StdEncoding.EncodeToString(hash[0:])
		return &credentialBody{id, encodedHash}, http.StatusOK
	}

	// return "Please specify user ID", http.StatusNotFound
	return "OK", http.StatusOK
}

func addHash(pathArgs []string, db *PasswordStore, r *http.Request) (interface{}, int) {
	// POST
	// return a job identifier (to find the password hash later)
	// wait 5 seconds
	// compute sha512 password hash and store it associated with the job id
	// store number of milliseconds it took the hash to complete
	userPassword := passwordBody{}
	inputErr := json.NewDecoder(r.Body).Decode(&userPassword)
	if inputErr != nil {
		return "Error decoding provided password", http.StatusInternalServerError
	}

	jobID := db.SavePassword([]byte(userPassword.Password))
	return credentialBody{UserID: jobID}, http.StatusOK

}

/*StatsHandler
A GET to /stats should accept no data;
it should return a JSON data structure for the total hash requests since server start
and the average time of a hash request in milliseconds, not including the 500ms wait.
*/
func getStats(pathArgs []string, db *PasswordStore, r *http.Request) (interface{}, int) {
	// GET
	// returns stats in the following example format
	/*
	  {
	    "total": 24,
	    "average": 123
	  }
	*/

	statsMessage := db.GetStats()
	stats := &statsBody{statsMessage.count, statsMessage.averageHashTime / time.Millisecond}
	return stats, http.StatusOK
}
