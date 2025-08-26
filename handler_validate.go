package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	// define the shape of your incoming JSON; The json:"body" tag tells Go how to map the JSON field 
	// to the struct field
	type parameters struct {
		Body string `json:"body"`
	}
	// define the shape of your outgoing JSON;
	type returnVals struct {
		// Valid bool `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	// create a decoder that reads from the request body:
	decoder := json.NewDecoder(r.Body)
	// create an empty parameters struct
	params := parameters{}
	// fill it with the JSON data; The &params passes a pointer so the decoder can modify the struct
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	// // if the chirp exceeds 140 characters, return an error:
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil) // nil = optional error parameter
		return
	}

	// Create a map where the keys are strings (our "bad words"), and the values are just empty structs:
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	// pass the chirp body and our map of badWords to the 'getCleanedBody' function:
	cleaned := getCleanedBody(params.Body, badWords)

	// If validation passes, respond with a JSON object containing the cleaned text:
		// (If validation passes, respond with a JSON object containing {"valid": true}:)
	respondWithJSON(w, http.StatusOK, returnVals{
		// Valid: true,
		CleanedBody: cleaned,
	})
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	//  takes the body string and splits it into a slice of 
	// smaller strings, using a space (" ") as the delimiter:
	words := strings.Split(body, " ")
	// iterate over the list of words:
	for i, word := range words {
		// convert the word to lowercase:
		loweredWord := strings.ToLower(word)
		// if the word is in badWords
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