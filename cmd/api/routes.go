package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("GET /v1/movies/{id}", app.showMovieHandler)
	mux.HandleFunc("POST /v1/movies", app.createMovieHandler)

	return app.recoverPanic(mux)
}
