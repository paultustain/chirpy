package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"example.com/m/v2/internal/auth"
	"example.com/m/v2/internal/database"
	"github.com/google/uuid"
)

type Parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

const ExpiresInSeconds = 3600

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
			ID:          userID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed.Bool,
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

	token, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Duration(ExpiresInSeconds)*time.Second,
	)

	if err != nil {
		respondWithError(w, 400, "failed to make jwt token:", err)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, 400, "failed to make refresh token:", err)
		return
	}
	_, err = cfg.db.CreateRefresh(r.Context(), database.CreateRefreshParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})

	if err != nil {
		respondWithError(w, 400, "failed to make refresh token:", err)
		return
	}
	respondWithJSON(
		w,
		200,
		response{
			User: User{
				ID:          user.ID,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
				Email:       user.Email,
				IsChirpyRed: user.IsChirpyRed.Bool,
			},
			Token:        token,
			RefreshToken: refreshToken,
		},
	)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 401, "failed to hash password: ", err)
		return
	}

	newUser, err := cfg.db.UpdateDetails(r.Context(), database.UpdateDetailsParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, 401, "failed to update database: ", err)
		return
	}

	respondWithJSON(w, 200,
		User{
			ID:        newUser.ID,
			CreatedAt: newUser.CreatedAt,
			UpdatedAt: newUser.UpdatedAt,
			Email:     newUser.Email,
		})

}

func (cfg *apiConfig) handlerUpdateMembership(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		UserID string `json:"user_id"`
	}
	type Params struct {
		Event string `json:"event"`
		Data  Data   `json:"data"`
	}

	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, "failed to get api key: ", err)
		return
	}

	if key != cfg.polkaSecret {
		respondWithError(w, 401, "incorrect api secret: ", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := Params{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 401, "failed to decode params: ", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, 401, "failed to parse userid: ", err)
		return
	}
	_, err = cfg.db.UpgradeMembership(r.Context(), userID)
	if err != nil {
		respondWithError(w, 404, "failed to upgrade database: ", err)
		return
	}

	w.WriteHeader(204)

}
