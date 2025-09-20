package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/craigbucher/learn-http-servers/internal/auth"
	"github.com/craigbucher/learn-http-servers/internal/database"
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
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	// create a new access token (a JWT):
	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,	// from .env
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	// generate a new refresh token and assigns it to refreshToken:
	refreshToken := auth.MakeRefreshToken()		// from auth.go

	// insert a new refresh-token row into the database for the user:
	// 	* The underscore ignores the returned record (you only care if it succeeded)
	//	* CreateRefreshToken is a sqlc-generated method
	// 	* On success, the row is created with created_at/updated_at set by your query/migration logic; 
	// 	  revoked_at should be NULL initially
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		// refresh tokens should expire after 60 days:
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	// send a successful JSON response with the public user fields (no password!)
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}
