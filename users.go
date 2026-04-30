package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
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
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error decoding paramters.", err)
		return
	}

	if params.Password == "" {
		RespondWithError(w, http.StatusNotAcceptable, "Password is required to create a user.", err)
		return
	}

	hashed_pw, err := auth.HashPassword(params.Password)
	if err != nil {
		RespondWithError(w, http.StatusNotAcceptable, "", err)
		return
	}

	if params.Email == "" {
		RespondWithError(w, http.StatusNotAcceptable, "Email is required to create a user.", err)
		return
	}

	dbuser, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_pw,
	})

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user.", err)
		return
	}

	user := User{
		ID:        dbuser.ID,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
		Email:     dbuser.Email,
	}

	RespondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error decoding paramters.", err)
		return
	}

	dbuser, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}

	valid, err := auth.CheckPasswordHash(params.Password, dbuser.HashedPassword)
	if err != nil || valid == false {
		RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}

	user := User{
		ID:        dbuser.ID,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
		Email:     dbuser.Email,
	}

	RespondWithJSON(w, http.StatusOK, user)
}
