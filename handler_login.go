package main

import (
	"encoding/json"
	"net/http"

	"github.com/craigbucher/learn-http-servers/internal/auth"
)

// Create a method on *apiConfig that handles HTTP requests to a login endpoint:
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Create a local struct to decode the JSON body:
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	// Create a local struct used to encode the JSON response:
	// It embeds a User type (Embedding means the User fields appear at the top level of the JSON)
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Call a DB method to fetch a user by email, passing the request context and the email from the 
	// parsed params. Returns the user record and an err:
	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	// verify the login password against the stored bcrypt hash:
	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	// send a successful JSON response with the public user fields (no password!)
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}
