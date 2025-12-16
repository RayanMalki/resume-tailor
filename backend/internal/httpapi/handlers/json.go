package handlers

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {

	//tells the clients we are returning JSON
	w.Header().Set("Content-Type", "applicatio/json")

	//set HTTP status code (200, 400, 500 etc etc)
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "failed to encode JSON", 500)
		return
	}

}

func writeError(w http.ResponseWriter, status int, msg string) {
	payload := map[string]string{
		"error": msg,
	}

	writeJSON(w, status, payload)
}
