package handlers

import "net/http"

func RootHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "welcome to the webiste"})
}
