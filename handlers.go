package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const getUserTimeout time.Duration = 10 * time.Second

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

	// for some reason, if we set route to /hash/ and call post on /hash, it uses the get method
	// but setting it to /hash stops get from accessing /hash/id
	// giving up for now. just post to /hash/
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

// getHash returns the base64 encoded hash matching the job it
func getHash(pathArgs []string, db *PasswordStore, r *http.Request) (interface{}, int) {
	/*example
	"WYPp76J/AShggbJxnW4PbgArjGOko6ZIAvJs9mh5dsiD5JUASXK45ktPC6L3cNszEcriajzZYrx+SmFAxFANeg=="
	*/
	if len(pathArgs) > 0 {
		userID := pathArgs[0]
		id, err := strconv.Atoi(userID)
		if err != nil {
			return "User ID must be an integer", http.StatusInternalServerError
		}

		hash, err := db.GetHash(id, getUserTimeout)
		if err != nil {
			return err.Error(), http.StatusNotFound
		}
		encodedHash := base64.StdEncoding.EncodeToString(hash[0:])
		return encodedHash, http.StatusOK
	}

	return "Must specify user ID", http.StatusNotFound
}

// addHash returns a job id, then notifies the password store manager to hash the password
func addHash(pathArgs []string, db *PasswordStore, r *http.Request) (interface{}, int) {
	/*example
	42
	*/
	var userPassword string
	inputErr := json.NewDecoder(r.Body).Decode(&userPassword)
	if inputErr != nil {
		err := inputErr.Error()
		return fmt.Sprint("Error decoding provided password", err), http.StatusInternalServerError
	}

	jobID, _ := db.SavePassword([]byte(userPassword))
	return jobID, http.StatusOK

}

// getStats returns total passwords hashed, and average time in milliseconds
func getStats(pathArgs []string, db *PasswordStore, r *http.Request) (interface{}, int) {
	/*example
	  {
	    "total": 24,
	    "average": 123
	  }
	*/
	stats := db.GetStats()
	stats.Average = stats.Average / time.Millisecond
	return stats, http.StatusOK
}
