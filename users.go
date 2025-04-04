package main

import (
	"encoding/json"
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
		respondWithError(w, 400, "failed to decode:", err)
		return
	}
	userID := uuid.New()
	_, err = cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:    userID,
		Email: params.Email,
	})

	if err != nil {
		respondWithError(w, 400, "failed to create user in db:", err)
		return
	}

	newUser := User{
		ID:        userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     params.Email,
	}

	dat, err := json.Marshal(newUser)
	if err != nil {
		respondWithError(w, 400, "failed to marshal:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}
