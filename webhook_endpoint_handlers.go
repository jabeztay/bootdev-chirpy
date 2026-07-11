package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jabeztay/bootdev-chirpy/internal"
	"github.com/jabeztay/bootdev-chirpy/internal/database"
)

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	rawPolkaKey, err := internal.GetAPIKey(r.Header)
	if err != nil || rawPolkaKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	type data struct {
		UserId string `json:"user_id"`
	}
	type parameters struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if params.Event == "user.upgraded" {
		userId, err := uuid.Parse(params.Data.UserId)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "user not found")
			return
		}

		_, err = cfg.dbQueries.UpdateChirpyRed(r.Context(), database.UpdateChirpyRedParams{IsChirpyRed: true, ID: userId})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusNotFound, "user not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "something went wrong")
			return
		}
	}

	type returnVals struct{}
	respBody := returnVals{}
	respondWithJson(w, http.StatusNoContent, respBody)
	return
}
