package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jabeztay/bootdev-chirpy/internal"
	"github.com/jabeztay/bootdev-chirpy/internal/database"
)

func (cfg *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}

	hashedPassword, err := internal.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashedPassword})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating a user")
		return
	}

	type returnVals struct {
		Id          string `json:"id"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}

	respBody := returnVals{
		Id:          user.ID.String(),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJson(w, http.StatusCreated, respBody)
	return
}

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	result, err := internal.CheckPasswordHash(params.Password, user.HashedPassword)

	if err != nil || !result {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	jwtExpiresInSeconds := 3600
	token, err := internal.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(jwtExpiresInSeconds)*time.Second)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating token")
		return
	}

	refreshToken := internal.MakeRefreshToken()
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: refreshToken, UserID: user.ID})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating refresh token")
		return
	}

	type returnVals struct {
		Id           string `json:"id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}

	respBody := returnVals{
		Id:           user.ID.String(),
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJson(w, http.StatusOK, respBody)
	return
}

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	refreshTokenResult, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil ||
		refreshTokenResult.RevokedAt.Valid ||
		refreshTokenResult.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jwtExpiresInSeconds := 3600
	token, err := internal.MakeJWT(refreshTokenResult.UserID, cfg.JWTSecret, time.Duration(jwtExpiresInSeconds)*time.Second)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating token")
		return
	}

	type returnVals struct {
		Token string `json:"token"`
	}

	respBody := returnVals{
		Token: token,
	}

	respondWithJson(w, http.StatusOK, respBody)
	return
}

func (cfg *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	type returnVals struct{}
	respBody := returnVals{}
	respondWithJson(w, http.StatusNoContent, respBody)
}

func (cfg *apiConfig) putUsersHandler(w http.ResponseWriter, r *http.Request) {
	accessToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userId, err := internal.ValidateJWT(accessToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request")
		return
	}

	hashedPassword, err := internal.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{Email: params.Email, HashedPassword: hashedPassword, ID: userId})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating a user")
		return
	}

	type returnVals struct {
		Id          string `json:"id"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}

	respBody := returnVals{
		Id:          user.ID.String(),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJson(w, http.StatusOK, respBody)
	return
}
