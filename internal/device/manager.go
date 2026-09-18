package device

import (
	"context"
	"errors"
	"sync"

	"github.com/coder/websocket"
)

type Manager struct {
	connMu  sync.RWMutex
	writeMu sync.Mutex

	conn *websocket.Conn
	auth Authenticator
}

func NewManager(auth Authenticator) *Manager {
	return &Manager{
		auth: auth,
	}
}

func (m *Manager) setConnection(conn *websocket.Conn) *websocket.Conn {
	m.connMu.Lock()
	defer m.connMu.Unlock() // Run this when function is about to return

	old := m.conn
	m.conn = conn

	return old
}

func (m *Manager) clearConnection(conn *websocket.Conn) {
	m.connMu.Lock()
	defer m.connMu.Unlock()

	// Only clear if this connection is still current
	// Don't delete a newer connection on accident
	if m.conn == conn {
		m.conn = nil
	}
}

func (m *Manager) Connected() bool {
	m.connMu.RLock()
	defer m.connMu.RUnlock()

	return m.conn != nil
}

func (m *Manager) SendCommand(
	ctx context.Context,
	command string,
) error {
	m.connMu.RLock()
	conn := m.conn
	m.connMu.RUnlock()

	if conn == nil {
		return errors.New("ESP32 is not connected")
	}

	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	return conn.Write(
		ctx,
		websocket.MessageText,
		[]byte(command),
	)
}
