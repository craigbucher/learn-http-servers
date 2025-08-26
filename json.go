package main

import (
	"encoding/json"
	"log"
	"net/http"
)

/*The function takes four parameters:

w: the HTTP response writer to send the response
code: the HTTP status code (like 400, 500, etc.)
msg: a user-friendly error message
err: the actual error object (which might be nil) */
func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		// log the error details to your server logs:
		log.Println(err)
	}
	if code > 499 {
		// If it's a server error (5XX codes), log that too:
		log.Printf("Responding with 5XX error: %s", msg)
	}
	// create a struct that will be converted to JSON. The json:"error" tag means the JSON will have 
	// an "error" field:
	type errorResponse struct {
		Error string `json:"error"`
	}
	// call respondWithJSON to send back a JSON response with the error message:
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

/*The function takes three parameters:

w: the HTTP response writer
code: the HTTP status code to send
payload: any data structure that can be converted to JSON (the interface{} means "any type")*/
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	// convert the payload (whatever struct or data you passed in) into JSON bytes:
	dat, err := json.Marshal(payload)
	if err != nil {
		// If there's an error, log it and send back a 500 status code:
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	// set the HTTP status code and writes the JSON data to the response
	w.WriteHeader(code)
	w.Write(dat)
}
