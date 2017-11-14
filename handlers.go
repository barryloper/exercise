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
func HashHandler(w http.ResponseWriter, r *http.Request) {}

/*StatsHandler
A GET to /stats should accept no data;
it should return a JSON data structure for the total hash requests since server start
and the average time of a hash request in milliseconds.
*/
func StatsHandler(w http.ResponseWriter, r *http.Request) {}
