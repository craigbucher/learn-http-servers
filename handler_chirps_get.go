package main

import (
	"net/http"
	"github.com/google/uuid"
	"sort"
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

	// Initialize authorID to the zero-value UUID (all zeros). This means “no author filter” by default:
	authorID := uuid.Nil
	// Read the optional author_id query param from the request URL as a string. If absent, it’s "":
	authorIDString := r.URL.Query().Get("author_id")
	// If authorIDString is not blank:
	if authorIDString != "" {
		// Convert the string to a UUID:
		authorID, err = uuid.Parse(authorIDString)
		// If parsing fails, it’s an invalid ID format:
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}
	}

	// set the default to ascending:
	sortDirection := "asc"
	// gets the sort query param from the URL:
	sortDirectionParam := r.URL.Query().Get("sort")
	// overrides the default only if the client asked for desc:
	if sortDirectionParam == "desc" {
		sortDirection = "desc"
	}

	// initialize an empty slice of your API’s Chirp type:
	chirps := []Chirp{}
	// loop over each DB record:
	for _, dbChirp := range dbChirps {
		// If the authorID is not nil and
		// the UserID of the chirp does not equal the authorID:
		if authorID != uuid.Nil && dbChirp.UserID != authorID {
			// skip this chirp:
			continue
		}
		// append a new Chirp (your response model) built from the DB row:
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	// sort the chirps slice in place:
	sort.Slice(chirps, func(i, j int) bool {
		if sortDirection == "desc" {
			// chirps[i].CreatedAt and chirps[j].CreatedAt are time.Time values
			// .After(other) returns true if the receiver is strictly later than other:
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

	// write: a successful JSON HTTP response:
	respondWithJSON(w, http.StatusOK, chirps)
}