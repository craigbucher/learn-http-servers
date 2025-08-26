package main

import "net/http"

// defines a function called handlerReadiness that can be used as an HTTP handler. It takes a 
// http.ResponseWriter (for writing the response) and an *http.Request (representing the incoming 
// HTTP request):
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// http.StatusText(http.StatusOK) evaluates to the string "OK".
	// w.Write([]byte(...)) writes the string "OK" as the body of the response:
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// This handler returns an HTTP 200 status code and the plain text "OK". It’s commonly used as a health 
// or readiness probe: other services or load balancers can make a request to this endpoint to check if 
// your server is alive and ready to handle traffic