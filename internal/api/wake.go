package api

import (
	"log"
	"net/http"
)

func (a *API) Wake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	if !a.authenticated(r) {
		http.Error(
			w,
			"Unauthorised",
			http.StatusUnauthorized,
		)
		return
	}

	err := a.devices.SendCommand(
		r.Context(),
		"wake",
	)

	if err != nil {
		http.Error(
			w,
			"Failed to send wake command",
			http.StatusServiceUnavailable,
		)
		return
	}

	log.Println("Wake sent")

	w.WriteHeader(http.StatusNoContent)
}
