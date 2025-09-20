package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/craigbucher/learn-http-servers/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request) {
	// define the shape of the JSON data your webhook endpoint expects to receive:
	type parameters struct {
		Event string `json:"event"`
		// nested struct that represents the "data" portion of the webhook payload:
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		}
	}

	// call the GetAPIKey function from the auth package and extracts the API key from the 
	// Authorization header:
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		// If there was an error extracting the API key,  responds with a 401 Unauthorized status:
		respondWithError(w, http.StatusUnauthorized, "Couldn't find api key", err)
		return
	}
	// if the extracted key doesn't match the one stored in .env, respond with a 401 
	// Unauthorized status:
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "API key is invalid", err)
		return
	}

	//  creates a new JSON decoder that reads from r.Body (the request body of the HTTP request):
	decoder := json.NewDecoder(r.Body)
	// creates an empty instance of your parameters struct 
	params := parameters{}
	// read the JSON from the request body and unmarshal (convert) it into your params struct:
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// check if the event type in the webhook payload is anything OTHER than "user.upgraded":
	if params.Event != "user.upgraded" {
		// if so, set the HTTP response status to 204 (No Content):
		w.WriteHeader(http.StatusNoContent)
		//  exit the function early, so no further processing happens:
		return
	}

	//  call the database function to update the user's is_chirpy_red field to true:
	// "_" = ignore the return value; we only care about the error
	_, err = cfg.db.UpgradeToChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		// if the database query didn't find any user with that UUID:
		if errors.Is(err, sql.ErrNoRows) {
			// respond with 404 Not Found (user doesn't exist):
			respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
			return
		}
		// for any other errors, respond with 500 Internal Server Error:
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}
	// Set the HTTP response status code to 204 (No Content):
	// (This tells the client (Polka) that the request was successfully processed)
	w.WriteHeader(http.StatusNoContent)
}
