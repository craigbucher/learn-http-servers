package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Declare a new custom type called TokenType as an alias for the built-in string type. 
// This means that a TokenType variable will behave just like a string under the hood, 
// but it provides type safety and readability:
// This helps prevent accidental misuse, as the compiler will warn you if you try to assign a plain 
// string where a TokenType is expected, or vice-versa, without an explicit conversion.
type TokenType string

// Declares a constant named TokenTypeAccess and explicitly assign the type TokenType, 
// even though its literal value "chirpy" is a string:
// (Constants are values that cannot be changed during the program's execution)
const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy"
)

// ErrNoAuthHeaderIncluded -
var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

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

// Add a MakeJWT function to your auth package:
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)
	// Create a new token:
	// 	* Use jwt.NewWithClaims:
	// 	* Use jwt.SigningMethodHS256 as the signing method:
	//	* Use jwt.RegisteredClaims as the claims:
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),		// Set the Issuer to "chirpy"
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),	// Set IssuedAt to the current time in UTC
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),	// Set ExpiresAt to the current time plus the expiration time (expiresIn)
		Subject:   userID.String(),	// Set the Subject to a stringified version of the user's id
	})
	// Use token.SignedString to sign the token with the secret key:
	return token.SignedString(signingKey)
}

// Add a ValidateJWT function to your auth package:
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// Initialize an instance of the jwt.RegisteredClaims struct to hold the standard claims defined 
	// in the JWT specification (things like Issuer, Subject, ExpiresAt, IssuedAt)
	// (you're using it to provide a container for jwt.ParseWithClaims to fill with the token's claims)
	claimsStruct := jwt.RegisteredClaims{}
	// Use the jwt.ParseWithClaims function to validate the signature of the JWT and extract the 
	// claims into a *jwt.Token struct:
	// (In essence, this single line of code tells the jwt library: "Here's a JWT string. Please 
	// parse it, use this empty claimsStruct as a template to fill with the token's payload data, 
	// and here's the secret key to use for verifying its signature. Tell me if it's valid and give 
	// me the parsed token object"):
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		// piece of code is a key lookup function (often called a Keyfunc in the context of the 
		// golang-jwt/jwt library). It's a special function that jwt.ParseWithClaims uses to get 
		// the secret key needed to verify the JWT's signature:
		// 	* It takes one parameter: token *jwt.Token
		//	* interface{}: This means the function can return any type of value (since keys used for signing can vary)
		//	* []byte(tokenSecret): take the tokenSecret string that was passed to your ValidateJWT 
		// 	function and converting it into a byte slice. This is the key that jwt.ParseWithClaims will 
		//  use to attempt to verify the token's signature
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}
	//  use the token.Claims interface to get access to the user's id from the claims 
	// (which should be stored in the Subject field):
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	// Get the issuer to validate the token:
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	// If issuer is not "chirpy", return an error:
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}
	// Return the id as a uuid.UUID:
	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return id, nil
}

// GetBearerToken -
// Auth information will come into our server in the Authorization header
func GetBearerToken(headers http.Header) (string, error) {
	// look for the Authorization header in the headers parameter:
	authHeader := headers.Get("Authorization")
	//  If the header doesn't exist, return an error:
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	// strip off the Bearer prefix and whitespace:
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}
	// return the TOKEN_STRING:
	return splitAuth[1], nil
}

// Add a func MakeRefreshToken() (string, error) function to your internal/auth package. It should 
// use the following to generate a random 256-bit (32-byte) hex-encoded string:
// 	* rand.Read to generate 32 bytes (256 bits) of random data from the crypto/rand package
// 	* hex.EncodeToString to convert the random data to a hex string
func MakeRefreshToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}

// GetAPIKey -
func GetAPIKey(headers http.Header) (string, error) {
	//  retrieve the value of the Authorization header from the HTTP request headers:
	authHeader := headers.Get("Authorization")
	// If authHeader is an empty string (meaning the header wasn't found), it returns an 
	// empty string and a custom error:
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	// split the Authorization header value by spaces:
	splitAuth := strings.Split(authHeader, " ")
	// Ensure there are at least 2 parts after splitting and
	// Verify that the first part is exactly "ApiKey":
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		// if not, return an error:
		return "", errors.New("malformed authorization header")
	}
	// return the second item in the slice, which is the API key:
	return splitAuth[1], nil
}