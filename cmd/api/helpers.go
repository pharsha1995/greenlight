package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, header http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for k, v := range header {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	
	return nil
}