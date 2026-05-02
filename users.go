package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
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
	Token     string    `json:"token"`
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

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, req *http.Request) {
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

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Error getting bearer token.", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Error validating token.", err)
		return
	}

	hashed_pw, err := auth.HashPassword(params.Password)
	if err != nil {
		RespondWithError(w, http.StatusNotAcceptable, "", err)
		return
	}

	dbuser, err := cfg.db.UpdateUser(req.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashed_pw,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user.", err)
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

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	expiresIn := time.Hour

	token, err := auth.MakeJWT(dbuser.ID, cfg.secret, time.Duration(expiresIn))
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error creating token.", err)
		return
	}

	refresh_token := auth.MakeRefreshToken()
	expires_at := time.Now().AddDate(0, 0, 60).UTC()

	_, err = cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     refresh_token,
		ExpiresAt: expires_at,
		RevokedAt: sql.NullTime{},
		UserID:    dbuser.ID,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error creating refresh token.", err)
		return
	}

	user := User{
		ID:        dbuser.ID,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
		Email:     dbuser.Email,
		Token:     token,
	}

	RespondWithJSON(w, http.StatusOK, response{
		User:         user,
		Token:        token,
		RefreshToken: refresh_token,
	})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, req *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	refresh_token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Error getting bearer token.", err)
		return
	}

	db_refresh_token, err := cfg.db.GetRefreshToken(req.Context(), refresh_token)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Failed to get refresh token.", err)
		return
	}

	//TODO: Consider moving this logic into the SQL select
	if time.Now().UTC().After(db_refresh_token.ExpiresAt) || db_refresh_token.RevokedAt.Valid {
		RespondWithError(w, http.StatusUnauthorized, "Refresh token expired or revoked.", err)
		return
	}

	expiresIn := time.Hour

	token, err := auth.MakeJWT(db_refresh_token.UserID, cfg.secret, time.Duration(expiresIn))
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error creating token.", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, req *http.Request) {
	refresh_token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Error getting bearer token.", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(req.Context(), refresh_token)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error revoking refresh token.", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(http.StatusText(http.StatusNoContent)))
}
