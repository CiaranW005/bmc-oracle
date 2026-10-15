package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CiaranW005/bmc-oracle/internal/device"
)

func (a *API) StatusEvents(w http.ResponseWriter, r *http.Request) {
	updates, unsubscribe := a.devices.SubscribeStatus()
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-accel-Buffering", "no")

	controller := http.NewResponseController(w)

	defer controller.SetWriteDeadline(time.Time{})

	write := func(message string) error {
		if err := controller.SetWriteDeadline(
			time.Now().Add(10 * time.Second),
		); err != nil {
			return err
		}

		if _, err := fmt.Fprint(w, message); err != nil {
			return err
		}

		return controller.Flush()
	}

	sendStatus := func(status device.Status) error {
		data, err := json.Marshal(status)
		if err != nil {
			return err
		}

		return write(fmt.Sprintf("data: %s\n\n", data))
	}

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case status := <-updates:
			if err := sendStatus(status); err != nil {
				return
			}

		case <-keepalive.C:
			// Keep connection alive without a UI update
			if err := write(": keepalive\n\n"); err != nil {
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}
