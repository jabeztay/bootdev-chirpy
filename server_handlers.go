package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jabeztay/bootdev-chirpy/internal/database"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden, this it not dev")
		return
	}

	cfg.dbQueries.ResetUsers(r.Context())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	userId, err := uuid.Parse(params.UserId)
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: params.Body, UserID: userId})
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
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

func (cfg *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, 500, "Something went wrong creating a user")
		return
	}

	type returnVals struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	respBody := returnVals{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	}

	respondWithJson(w, 201, respBody)
	return
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong fetching chirps")
		return
	}

	type returnVals struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserId    string `json:"user_id"`
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

	respondWithJson(w, 200, respBody)
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

	respondWithJson(w, 200, respBody)
}
