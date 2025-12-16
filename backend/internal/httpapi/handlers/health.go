package handlers

import (
	"net/http"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "health is ok"})

}
