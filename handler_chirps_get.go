package main

import (
	"net/http"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	// Reads the {chirpID} path parameter from the request URL using Go’s http.Request.PathValue:
	// from: mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerChirpsGet)
	// Because the route uses {chirpID}, calling r.PathValue("chirpID") returns that segment from the URL
	chirpIDString := r.PathValue("chirpID")
	// Validate and convert that string into a uuid.UUID:
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	// Call the database method to fetch a single chirp (from sql/queries/chirps) with chirpID:
	// Pass the request context so timeouts/cancellation propagate:
	// Return dbChirp (a single chirp) and err:
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}
	
	// Create a new value with fields copied from dbChirp:
	// Serializes that value to JSON, sets status 200, writes to the ResponseWriter:
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
	})
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	// Call the database method to fetch all chirps (from sql/queries/chirps):
	// Pass the request context so timeouts/cancellation propagate:
	// Return dbChirps (slice of chirps) and err:
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	// initialize an empty slice of your API’s Chirp type:
	chirps := []Chirp{}
	// loop over each DB record:
	for _, dbChirp := range dbChirps {
		// append a new Chirp (your response model) built from the DB row:
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	// write: a successful JSON HTTP response:
	respondWithJSON(w, http.StatusOK, chirps)
}