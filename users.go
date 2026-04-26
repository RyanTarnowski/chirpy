package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error decoding paramters.", err)
		return
	}
	if params.Email == "" {
		RespondWithError(w, http.StatusNotAcceptable, "Email is required to create a user.", err)
	}
	dbuser, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to create user.", err)
	}

	user := User{
		ID:        dbuser.ID,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
		Email:     dbuser.Email,
	}

	RespondWithJSON(w, http.StatusCreated, user)
}
