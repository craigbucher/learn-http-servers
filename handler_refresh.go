package main

import (
	"net/http"
	"time"

	"github.com/craigbucher/learn-http-servers/internal/auth"
)

// define a method called handlerRefresh that:
//	* Takes two parameters that are standard for HTTP handlers in Go
// 	* defines a local struct type called response that will be used to format the JSON response
func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	// extract and validate the refresh token from the HTTP request headers:
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	// look-up up the user associated with the refresh token in the database:
	// r.Context() - the request context (used for timeouts, cancellation, etc.)
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}
	// call the MakeJWT function from your auth package:
	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,	// from .env
		time.Hour,		// expiration duration = 1 hour
	)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}
	// create an instance of the response struct, setting the Token field to the JWT created above:
	// respond with a 200 code = http.StatusOK
	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

// define a method called handlerRevoke that:
//	* Takes two parameters that are standard for HTTP handlers in Go
func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	// Read the refresh token from Authorization: Bearer <token> in r.Header:
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}
	// Mark that token as revoked in the DB (set revoked_at and update updated_at):
	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}
	// Write status 204 with no body:
	w.WriteHeader(http.StatusNoContent)
}
