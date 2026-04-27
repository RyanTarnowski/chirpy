package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error decoding paramters.", err)
		return
	}

	if params.Body == "" {
		RespondWithError(w, http.StatusNotAcceptable, "Body is required to create a chirp.", err)
		return
	}

	validChirp, err := validateChirp(params.Body)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	res, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   validChirp,
		UserID: params.UserID,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create chirp.", err)
		return
	}

	chirp := Chirp{
		ID:        res.ID,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
		Body:      res.Body,
		UserId:    res.UserID,
	}

	RespondWithJSON(w, http.StatusCreated, chirp)
}

func validateChirp(body string) (string, error) {
	const maxChripLen = 140

	if len(body) > maxChripLen {
		return "", errors.New("Chirp is too long")
	}

	validChirp := profanityScrubber(body)

	return validChirp, nil
}

func profanityScrubber(chirp string) string {
	const redact = "****"
	badWords := []string{
		"kerfuffle", "sharbert", "fornax",
	}

	for word := range strings.SplitSeq(chirp, " ") {
		if slices.Contains(badWords, strings.ToLower(word)) {
			chirp = strings.ReplaceAll(chirp, word, redact)
		}
	}

	return chirp
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	//TODO: Get all chirps from the db and return them in asc order by create date
	//RespondWithJSON(w, http.StatusCreated, chirp)
}
