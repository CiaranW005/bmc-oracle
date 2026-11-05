package device

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/CiaranW005/go-protocol"
)

func (m *Manager) StartTelemetry(ctx context.Context) {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.pollTelemetry(ctx)

		case <-ctx.Done():
			return
		}
	}
}

func (m *Manager) pollTelemetry(ctx context.Context) {
	message := protocol.NewMessage(protocol.TypeTelemetry, nil)

	packet, err := m.SendAndWait(
		ctx,
		message,
		protocol.RetryPolicy{
			Timeout:    2 * time.Second,
			MaxRetries: 0,
		},
		3*time.Second,
	)

	if err != nil {
		log.Printf("Failed to poll for telemetry")
		m.publishStatus(false, false)
		return
	}

	if packet.Type != protocol.TypeTelemetry {
		log.Printf("Unexpected telemetry response type: %s", packet.Type)
		m.publishStatus(false, false)
		return
	}

	raw, err := json.Marshal(packet.Body)
	if err != nil {
		log.Printf("Invalid telemtry body: %v", err)
		m.publishStatus(false, false)
		return
	}

	var body struct {
		PC *bool `json:"pc"` // A pointer distinguishes a valid false value from a missing or null pc field.
	}

	if err := json.Unmarshal(raw, &body); err != nil || body.PC == nil {
		log.Printf("Invalid telemtry reponse; missing or invalid pc field")
		m.publishStatus(false, false)
		return
	}

	// A valid telemetry response confirms BMC availability.
	// PC availability comes from the ESP32's report of the PC agent.
	m.publishStatus(true, *body.PC)
}
