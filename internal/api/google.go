package api

import (
	"encoding/json"
	"net/http"
)

type googleLoginRequest struct {
	Credential string `json:"credential"`
}

func (a *API) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*10240)

	var req googleLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := a.googleAuth.Verify(r.Context(), req.Credential); err != nil {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	if err := a.auth.StartSession(w); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
