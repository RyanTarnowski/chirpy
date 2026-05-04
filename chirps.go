package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"
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

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Error getting bearer token.", err)
		return
	}

	params.UserID, err = auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Error validating token.", err)
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

func authorIDFromRequest(r *http.Request) (uuid.UUID, error) {
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString == "" {
		return uuid.Nil, nil
	}
	authorID, err := uuid.Parse(authorIDString)
	if err != nil {
		return uuid.Nil, err
	}
	return authorID, nil
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	authorID, err := authorIDFromRequest(req)
	sortDirection := req.URL.Query().Get("sort")
	dbChirps := []database.Chirp{}

	if authorID != uuid.Nil {
		dbChirps, err = cfg.db.GetChirpsByAuthor(req.Context(), authorID)
	} else {
		dbChirps, err = cfg.db.GetChirps(req.Context())
	}
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get chirps", err)
		return
	}

	sort.Slice(dbChirps, func(i, j int) bool {
		if sortDirection == "desc" {
			return dbChirps[i].CreatedAt.After(dbChirps[j].CreatedAt)
		}

		return dbChirps[i].CreatedAt.Before(dbChirps[j].CreatedAt)
	})

	chirps := []Chirp{}
	for _, c := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserId:    c.UserID,
		})
	}

	RespondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, req *http.Request) {
	reqID := req.PathValue("chirpID")

	chirpID, err := uuid.Parse(reqID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to parse chirp ID", err)
		return
	}

	res, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Failed to get chirp", err)
		return
	}

	chirp := Chirp{
		ID:        res.ID,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
		Body:      res.Body,
		UserId:    res.UserID,
	}

	RespondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	reqID := req.PathValue("chirpID")

	chirpID, err := uuid.Parse(reqID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to parse chirp ID.", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Failed to get chirp.", err)
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

	if chirp.UserID != userID {
		RespondWithError(w, http.StatusForbidden, "Only the author of a chirp can delete it.", err)
		return
	}

	err = cfg.db.DeleteChirpByID(req.Context(), chirpID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete chirp.", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(http.StatusText(http.StatusNoContent)))
}
