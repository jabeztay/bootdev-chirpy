package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jabeztay/bootdev-chirpy/internal"
	"github.com/jabeztay/bootdev-chirpy/internal/database"
)

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := internal.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userId, err := internal.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: params.Body, UserID: userId})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type returnVals struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserId    string `json:"user_id"`
	}
	respBody := returnVals{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
	}
	respondWithJson(w, 201, respBody)
	return
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")

	var chirps []database.Chirp
	var err error

	if authorId == "" {
		chirps, err = cfg.dbQueries.GetChirps(r.Context())
	} else {
		userId, err := uuid.Parse(authorId)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid userid")
		}
		chirps, err = cfg.dbQueries.GetChirpsByUser(r.Context(), userId)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong fetching chirps")
		return
	}

	type returnVals struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserId    string `json:"user_id"`
	}

	sort := r.URL.Query().Get("sort")

	if strings.ToLower(sort) == "desc" {
		slices.Reverse(chirps)
	}

	respBody := make([]returnVals, 0, len(chirps))
	for _, chirp := range chirps {
		chirpParsed := returnVals{
			Id:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
			Body:      chirp.Body,
			UserId:    chirp.UserID.String(),
		}
		respBody = append(respBody, chirpParsed)
	}

	respondWithJson(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}
	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	type returnVals struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserId    string `json:"user_id"`
	}

	respBody := returnVals{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
	}

	respondWithJson(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := internal.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userId, err := internal.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "User is not author")
		return
	}

	err = cfg.dbQueries.DeleteChirpById(r.Context(), chirpId)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unexpected error")
	}

	type returnVals struct{}
	respBody := returnVals{}
	respondWithJson(w, http.StatusNoContent, respBody)
}
