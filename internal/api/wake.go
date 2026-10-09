package api

import (
	"log"
	"net/http"
)

func (a *API) sendCommand(w http.ResponseWriter, r *http.Request, command string) {
	err := a.devices.SendCommand(r.Context(), command)
	if err != nil {
		http.Error(
			w,
			"Failed to send "+command+" command",
			http.StatusServiceUnavailable,
		)
		return
	}

	log.Printf("Command sent: %s\n", command)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) Wake(w http.ResponseWriter, r *http.Request) {
	a.sendCommand(w, r, "wake")
}

func (a *API) Reset(w http.ResponseWriter, r *http.Request) {
	a.sendCommand(w, r, "reset")
}

func (a *API) PowerOff(w http.ResponseWriter, r *http.Request) {
	a.sendCommand(w, r, "power_off")
}
