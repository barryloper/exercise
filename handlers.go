package main

import (
	"net/http"
)

/*HashHandler
  A POST to /hash should accept a password; it should return a job identifier immediate;
  it should then wait 5 seconds and compute the password hash.
  The hashing algorithm should be SHA512.

  A GET to /hash should accept a job identifier;
  it should return the base64 encoded password hash for the corresponding POST request.
*/
func HashHandler(w http.ResponseWriter, r *http.Request) {
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
}
