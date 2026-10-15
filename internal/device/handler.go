package device

import (
	"context"
	"log"
	"net/http"

	"github.com/CiaranW005/go-protocol"
	"github.com/coder/websocket"
)

func (m *Manager) Handler(w http.ResponseWriter, r *http.Request) {
	if !m.auth.Valid(r) {
		log.Println("Device authentication failed")
		log.Printf("Authorization header present: %t", r.Header.Get("Authorization") != "")

		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	log.Println("Device authenticated")

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Websocket accept failed: %v", err)
		return
	}

	log.Println("WebSocket connected")

	// if another ESP32 connection exists, replace it with a this new one
	old := m.setConnection(conn)

	if old != nil && old != conn {
		_ = old.Close(
			websocket.StatusNormalClosure,
			"Replaced by a new device connection",
		)
	}

	// When this handler eventually exits, remove this connection if its still current
	defer m.clearConnection(conn)

	log.Println("ESP32 connected")

	m.readLoop(conn)
}

func (m *Manager) readLoop(conn *websocket.Conn) {
	for {
		_, data, err := conn.Read(context.Background())
		if err != nil {
			log.Println("ESP32 disconnected:", err)
			return
		}

		packet, err := protocol.DeserialisePacket(data)
		if err != nil {
			log.Printf("Invalid pakcet: %v\n", err)
			continue
		}

		switch p := packet.(type) {
		case protocol.Acknowledgement:
			if !m.tracker.Acknowledgement(p) {
				log.Printf("Unexpected acknowledgement: %s\n", p.ID)
			}

		case protocol.Message:
			if m.dedup.IsDuplicate(p.ID) {
				log.Printf("Duplicate message: %s\n", p.ID)

				ack := protocol.NewAcknowledgement(
					p.ID,
					protocol.AckOk,
				)

				if err := m.SendPacket(context.Background(), ack); err != nil {
					log.Printf("Failed to acknowledge duplicate: %v\n", err)
				}

				continue
			}

			log.Printf("Recieved message: %s\n", p.Type)

			ack := protocol.NewAcknowledgement(
				p.ID,
				protocol.AckOk,
			)

			if err := m.SendPacket(context.Background(), ack); err != nil {
				log.Printf("Failed to acknowledge message: %v\n", err)
			}

			if p.ReplyTo != "" {
				if !m.responses.Resolve(p) {
					log.Printf("Unexpected or late response: %s", p.ReplyTo)
				}
				continue
			}
		}
	}
}
