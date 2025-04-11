package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"example.com/m/v2/internal/auth"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	cfg.db.ResetUsers(r.Context())
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset fileserverHits"))
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "couldn't get bearer token: ", err)
	}
	token, err := cfg.db.GetRefresh(
		r.Context(),
		bearer,
	)
	if err != nil {
		respondWithError(w, 401, "couldn't get token from db: ", err)
		return
	}
	if token.RevokedAt.Valid {
		respondWithError(w, 401, "token revoked: ", errors.New(""))
		return
	}
	if token.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "token expired: ", errors.New(""))
		return
	}

	newToken, err := auth.MakeJWT(
		token.UserID,
		cfg.jwtSecret,
		time.Duration(time.Hour),
	)

	if err != nil {
		respondWithError(w, 400, "failed to make jwt token: ", err)
		return
	}

	respondWithJSON(w, 200, response{
		Token: newToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "couldn't get bearer token: ", err)
		return
	}

	err = cfg.db.RevokeToken(r.Context(), bearer)
	if err == sql.ErrNoRows {
		respondWithError(w, 404, "no token found to revoke: ", err)
		return
	}
	if err != nil {
		respondWithError(w, 401, "couldn't revoke token: ", err)
		return
	}

	w.WriteHeader(204)
}
