package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"example.com/m/v2/internal/database"
	"github.com/google/uuid"
)

type Parameters struct {
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("failed to decode: %s", err))
		return
	}

	_, err = cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:    uuid.New(),
		Email: params.Email,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("failed to create user in db: %s", err))
		return
	}

	newUser := User{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     params.Email,
	}

	dat, err := json.Marshal(newUser)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("failed to marshal: %s", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}
