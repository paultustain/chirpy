package main

import (
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("."))
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	mux.Handle("/", fs)

	server.ListenAndServe()

}
