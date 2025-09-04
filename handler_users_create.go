package main

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/google/uuid"
	"github.com/craigbucher/learn-http-servers/internal/auth"
	"github.com/craigbucher/learn-http-servers/internal/database"
)

// I created a User struct in my main package. When the database package returns a database.User, I map 
// it to my main package's User struct before marshalling it to JSON so that I can control the JSON keys:
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` 		// json means don't unmarshal from JSON, don't marshal to JSON (ignore)
}

//  create a method on apiConfig that handles POST /api/users:
func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// create a struct for the shape of the expected JSON request body:
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	// create a struct for the shape of the JSON response; it embeds your local User type (created above)
	//  so its fields (id, created_at, etc.) are included in the output:
	type response struct {
		User
	}

	// create a JSON decoder that reads from the HTTP request body:
	decoder := json.NewDecoder(r.Body)
	// allocate a zero-value params struct:
	params := parameters{}
	// parse the JSON body into params:
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// calls your bcrypt-based helper to turn the raw password into a secure hash. It returns the 
	// hash string and an error:
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Call the SQLC-generated CreateUser with the request context and the email from the parsed body:
		// r.Context(): ties the DB call to the HTTP request (cancels on timeout/abort)
		// On success, user holds the newly created row (id, timestamps, email)
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	// Set HTTP status to 201 Created
	// Write a JSON body shaped like response, containing a User built from the DB user:
	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}
