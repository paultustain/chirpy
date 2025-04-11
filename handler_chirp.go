package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"example.com/m/v2/internal/auth"
	"example.com/m/v2/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func replaceWords(body string) string {
	words := strings.Split(body, " ")

	for i := range words {
		if (strings.ToLower(words[i]) == "kerfuffle") || (strings.ToLower(words[i]) == "sharbert") || (strings.ToLower(words[i]) == "fornax") {
			words[i] = "****"
		}

	}
	return strings.Join(words, " ")
}

func validateChirp(chirp string) (string, error) {

	if len(chirp) > 140 {
		return "", errors.New("chirp too long")
	}

	return replaceWords(chirp), nil

}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {

	type Params struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(
			w,
			401,
			"failed to get bearer token",
			err,
		)
	}
	userID, err := auth.ValidateJWT(token, os.Getenv("SECRET"))

	if err != nil {
		respondWithError(
			w,
			401,
			"failed to get validate token",
			err,
		)
	}

	decoder := json.NewDecoder(r.Body)
	params := Params{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong:", err)
		return
	}

	cleanedChirp, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to validate:", err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedChirp,
		UserID: userID,
	})

	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"failed to create chirp in db: ",
			err,
		)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpRetrieve(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(param)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to parse chirp ID: ", err)
		return

	}
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "failed to get chirp", err)
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(param)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to parse chirp ID: ", err)
		return

	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)

	if err != nil {
		respondWithError(w, 404, "chirp id not found: ", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(
			w,
			401,
			"failed to get bearer token: ",
			err,
		)
		return
	}

	userID, err := auth.ValidateJWT(token, os.Getenv("SECRET"))
	if err != nil {
		respondWithError(
			w,
			401,
			"failed to get validate token: ",
			err,
		)
		return
	}

	if dbChirp.UserID != userID {
		respondWithError(
			w,
			403,
			"cannot delete other users chirp: ",
			err,
		)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), dbChirp.ID)
	if err != nil {
		respondWithError(
			w,
			401,
			"failed to delete chirp: ",
			err,
		)
		return
	}

	w.WriteHeader(204)

}
