package main

import "net/http"

// declares a method named handlerReset on the *apiConfig struct. It’s designed to be used as an HTTP 
// handler, so it receives a http.ResponseWriter and an *http.Request
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	// This line resets the hit counter by storing the value 0 in fileserverHits. The Store method is 
	// safe for concurrent use, which is crucial since your server may handle multiple requests at once:
	cfg.fileserverHits.Store(0)
	// explicitly set the HTTP status code to 200 (OK), indicating success:
	w.WriteHeader(http.StatusOK)
	// send the text "Hits reset to 0" as the response body. This lets clients know the action was performed
	w.Write([]byte("Hits reset to 0"))
}

// This handler sets the counter to zero for all subsequent requests.
// It responds with a 200 status and a confirmation message.
// It’s safe to use even under heavy concurrency because of the atomic operation.