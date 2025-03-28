package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type chirpBody struct {
	Body string `json:"body"`
}

type validateResponse struct {
	Valid bool   `json:"valid,omitempty"`
	Error string `json:"error,omitempty"`
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
	w.WriteHeader(400)
	w.Write(dat)
}

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	chirp := chirpBody{}

	err := decoder.Decode(&chirp)

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("something went wrong: %s", err))
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "chirp too long")
		return
	}

	resp := &validateResponse{
		Valid: true,
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("failed to marshal: %s", err))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)

}
