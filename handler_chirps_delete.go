package main

import (
	"net/http"

	"github.com/craigbucher/learn-http-servers/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	//  extract and validate a chirpID from a URL path:
	chirpIDString := r.PathValue("chirpID")
	// convert the chirpIDString (which is a string) into a uuid.UUID type:
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		// If the chirp is not found, return a 404 status code:
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}
	// extract a "Bearer" token from the Authorization header of the incoming HTTP request :
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	// validate the JSON Web Token (JWT) extracted above:
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// fetch the chirp by ID from the database:
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		// If the chirp is not found, return a 404 status code:
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}
	// Compare the chirp’s author (dbChirp.UserID) to the authenticated user’s ID (userID):
	if dbChirp.UserID != userID {
		// If not the same, return HTTP 403 Forbidden because the user is authenticated but 
		// not allowed to delete this resource:
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp", err)
		return
	}
	// Call the DB layer to delete the chirp with that ID:
	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}
	// If the chirp is deleted successfully, return a 204 status code:
	w.WriteHeader(http.StatusNoContent)
}
