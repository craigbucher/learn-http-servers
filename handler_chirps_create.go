package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"github.com/craigbucher/learn-http-servers/internal/database"
	"github.com/google/uuid"
)

// Create a struct definition for a “chirp” resource: 
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	// define the shape of your incoming JSON; The json:"body" tag tells Go how to map the JSON field 
	// to the struct field:
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	// create a decoder that reads from the request body:
	decoder := json.NewDecoder(r.Body)
	// create an empty parameters struct:
	params := parameters{}
	// fill it with the JSON data; The &params passes a pointer so the decoder can modify the struct:
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Call the 'validateChirp' method on the parameter/chirp body:
	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Create a chirp in the database and handle any errors:
	// cfg.db.CreateChirp = call the DB method 'CreateChirp' (from chirps.sql, created by sqlc) to insert a row:
	// database.CreateChirpParams = parameters to insert into the database:
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,		// validated/sanitized chirp body
		UserID: params.UserID,	// the author’s UUID
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	// call the 'respondWithJason' method from json.go:
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func validateChirp(body string) (string, error) {
	// if the chirp exceeds 140 characters, return an error:
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}
	// Create a map where the keys are strings (our "bad words"), and the values are just empty structs:
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	// pass the chirp body and our map of badWords to the 'getCleanedBody' function:
	cleaned := getCleanedBody(body, badWords)
	// Return the cleaned body and no error:
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	// takes the body string and splits it into a slice of 
	// smaller strings, using a space (" ") as the delimiter:
	words := strings.Split(body, " ")
	// iterate over the list of words:
	for i, word := range words {
		// convert the word to lowercase:
		loweredWord := strings.ToLower(word)
		// if the word is in badWords:
		if _, ok := badWords[loweredWord]; ok {
			// replace it with 4 asterixes:
			words[i] = "****"
		}
	}
	// take the slice of words and joins them back together into a single string, using a space (" ") 
	// as the separator between them:
	cleaned := strings.Join(words, " ")
	// return the cleaned string, which now has any profane words replaced with asterisks:
	return cleaned
}
