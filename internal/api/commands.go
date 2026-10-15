package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/CiaranW005/go-protocol"
)

func (a *API) sendCommand(
	w http.ResponseWriter,
	r *http.Request,
	command protocol.MessageType,
	timeout time.Duration,
) {
	writeResult := func(status int, state, message string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if err := json.NewEncoder(w).Encode(struct {
			Command protocol.MessageType `json:"command"`
			State   string               `json:"state"`
			Error   string               `json:"error,omitempty"`
		}{
			Command: command,
			State:   state,
			Error:   message,
		}); err != nil {
			log.Printf("Failed to write command response: %v", err)
		}
	}

	message := protocol.NewMessage(command, nil)

	response, err := a.devices.SendAndWait(
		r.Context(),
		message,
		protocol.RetryPolicy{
			Timeout:    3 * time.Second,
			MaxRetries: 0,
		},
		timeout,
	)

	if errors.Is(err, protocol.ErrResponseTimeout) {
		switch command {
		case protocol.TypeReset, protocol.TypePowerOff:
			writeResult(http.StatusOK, "no_failure_reported", "")
		default:
			writeResult(
				http.StatusGatewayTimeout,
				"failed",
				"PC did not come online before the deadline",
			)
		}
		return
	}

	if err != nil {
		log.Printf("Command %s failed: %v", command, err)
		writeResult(
			http.StatusServiceUnavailable,
			"failed",
			"Could not complete command",
		)
		return
	}

	// Message.Body is decoded as a map when received over the wire.
	// Convert it into the expected response body.
	data, err := json.Marshal(response.Body)
	if err != nil {
		writeResult(http.StatusBadGateway, "failed", "Invalid response body")
		return
	}

	var result struct {
		Success *bool  `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(data, &result); err != nil ||
		result.Success == nil {
		writeResult(http.StatusBadGateway, "failed", "Invalid response body")
		return
	}

	if !*result.Success {
		if result.Error == "" {
			result.Error = "Device reported command failure"
		}
		writeResult(http.StatusBadGateway, "failed", result.Error)
		return
	}

	if command == protocol.TypeWake {
		writeResult(http.StatusOK, "online", "")
		return
	}

	writeResult(http.StatusOK, "completed", "")
}

func (a *API) Wake(w http.ResponseWriter, r *http.Request) {
	a.sendCommand(w, r, protocol.TypeWake, 20*time.Second)
}

func (a *API) Reset(w http.ResponseWriter, r *http.Request) {
	a.sendCommand(w, r, protocol.TypeReset, 5*time.Second)
}

func (a *API) PowerOff(w http.ResponseWriter, r *http.Request) {
	a.sendCommand(w, r, protocol.TypePowerOff, 5*time.Second)
}
