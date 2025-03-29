package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type chirpBody struct {
	Body string `json:"body"`
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
		Valid:       true,
		CleanedBody: replaceWords(chirp.Body),
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
