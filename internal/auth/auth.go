package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Hash the password using the bcrypt.GenerateFromPassword function:
func HashPassword(password string) (string, error) {
	// hashe the password using bcrypt with a sensible work factor. Return a byte slice and an error:
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	// convert the hash bytes to a string and return it:
	return string(dat), nil
}

// Use the bcrypt.CompareHashAndPassword function to compare the password that the user entered 
// in the HTTP request with the password that is stored in the database:
// the full stored hash includes the algorithm, cost, and salt
func CheckPasswordHash(password, hash string) error {
	// re-hashe the password using the parameters embedded in hash and compare it to the raw password bytes:
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

/* In bcrypt, the “work factor” (cost) controls how slow hashing is. Higher cost = more CPU time = stronger against brute force.

Default in Go: bcrypt.DefaultCost (currently 10)
Typical choices today: 10–12 for web backends
Pick as high as you can while keeping login/signup latency acceptable (e.g., <100–200 ms per hash on your servers)
Benchmark on your deployment and set a fixed cost accordingly. */