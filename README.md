# exercise
building a simple api in golang


build:
```
go build
./exercise
```

optional:
```
./exercise -address <listen IP> -port <listen port>
```

# endpoints

All endpoints accept and return JSON

POST /hash/

returns a new hash ID and begins computing a hash.  
Wait at least 5 seconds to try to retrieve the hash  
KNOWN BUG: the trailing slash is mandatory here

example response:
```
"42"
```

GET /hash/{hashID}  

retrieves a base64 encoded hash associated with the ID specified

example response:  
```
"k6eKiHGXBb2LpwJM2UsgdFd+42X7Xn8kD3Q4caIBHSY/fiTs/3g3kDDomdA7P4fNY6oNRRa2ZgtEmdjAUK2sAQ=="
```

GET /stats

returns the number of hashes computed, and the average time to hash in milliseconds

example response:

```
{
    total: 42,
    average: 10
}
```

