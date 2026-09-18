package device

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func (m *Manager) Handler(w http.ResponseWriter, r *http.Request) {
	if !m.auth.Valid(r) {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("Websocket accept failed:", err)
		return
	}

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
		_, msg, err := conn.Read(context.Background())
		if err != nil {
			log.Println("ESP32 disconnected:", err)
			return
		}

		log.Printf("ESP32: %s\n", string(msg))
	}
}
