package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"example.com/m/v2/internal/database"
	"github.com/google/uuid"
)

type chirpBody struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type validateResponse struct {
	Valid       bool   `json:"valid,omitempty"`
	Error       string `json:"error,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	fmt.Println(msg)
	resp := &validateResponse{
		Error: msg,
	}
	dat, respErr := json.Marshal(resp)
	if respErr != nil {
		fmt.Printf("failed to marshal response: %s", respErr)
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func validateChirp(chirp string) (validateResponse, error) {

	if len(chirp) > 140 {
		return validateResponse{}, errors.New("chirp too long")
	}

	resp := &validateResponse{
		Valid:       true,
		CleanedBody: replaceWords(chirp),
	}

	return *resp, nil

}

func (cfg *apiConfig) handlerChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirpOriginal := chirpBody{}
	err := decoder.Decode(&chirpOriginal)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("something went wrong: %s", err))
		return
	}

	chirp, err := validateChirp(chirpOriginal.Body)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("failed to validate: %s", err))
		return
	}
	resp := Chirp{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      replaceWords(chirp.CleanedBody),
		UserID:    chirpOriginal.UserID,
	}

	dat, err := json.Marshal(&resp)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("failed to marshal: %s", err))
	}

	cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:   uuid.New(),
		Body: replaceWords(chirp.CleanedBody),
		UserID: uuid.NullUUID{
			UUID:  chirpOriginal.UserID,
			Valid: true,
		},
	})

	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}
