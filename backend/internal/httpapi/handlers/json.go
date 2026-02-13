package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {

	//tells the clients we are returning JSON
	w.Header().Set("Content-Type", "application/json")

	//set HTTP status code (200, 400, 500 etc etc)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Headers are already sent so the best we can do is log; the status code
		// was already written by WriteHeader above.
		return
	}

}

func writeError(w http.ResponseWriter, status int, msg string) {
	payload := map[string]string{
		"error": msg,
	}

	writeJSON(w, status, payload)
}

func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
