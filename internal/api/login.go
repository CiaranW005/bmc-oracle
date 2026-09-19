package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type loginRequest struct {
	Password string `json:"password"`
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login handler reached")
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	if !a.auth.CheckPassword(req.Password) {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	err := a.auth.StartSession(w)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	a.auth.EndSession(w, r)
	w.WriteHeader(http.StatusNoContent)
}
