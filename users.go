package main

import (
	"encoding/json"
	"net/http"
	"time"

	"example.com/m/v2/internal/auth"
	"example.com/m/v2/internal/database"
	"github.com/google/uuid"
)

type Parameters struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 400, "failed to decode:", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, "failed to hash password", err)
		return
	}

	userID := uuid.New()
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {

		respondWithError(w, 400, "failed to create user in db:", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        userID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 400, "failed to decode:", err)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 400, "failed to get user:", err)
		return
	}
	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "using an incorrect password", err)
		return
	}

	if params.ExpiresInSeconds <= 0 || params.ExpiresInSeconds > 3600 {
		params.ExpiresInSeconds = 3600 // Default to 1 hour if out of bounds or omitted
	}

	token, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Duration(params.ExpiresInSeconds)*time.Second,
	)

	if err != nil {
		respondWithError(w, 400, "failed to make jwt token:", err)
		return
	}

	respondWithJSON(
		w,
		200,
		response{
			User: User{
				ID:        user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email:     user.Email,
			},
			Token: token,
		},
	)
}
