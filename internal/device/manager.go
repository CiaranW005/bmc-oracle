package device

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CiaranW005/go-protocol"
	"github.com/coder/websocket"
)

type Status struct {
	BMC      bool      `json:"bmc"`
	PC       bool      `json:"pc"`
	LastSeen time.Time `json:"lastSeen"`
}

type Manager struct {
	connMu  sync.RWMutex
	writeMu sync.Mutex

	conn *websocket.Conn
	auth Authenticator

	tracker   *protocol.AckTracker
	responses *protocol.ResponseTracker
	dedup     *protocol.Deduplicator

	statusMu    sync.RWMutex
	status      Status
	subscribers map[chan Status]struct{}
}

func NewManager(auth Authenticator) *Manager {
	m := &Manager{
		auth:        auth,
		dedup:       protocol.NewDeduplicator(5 * time.Minute),
		responses:   protocol.NewResponseTracker(),
		subscribers: make(map[chan Status]struct{}),
	}

	m.tracker = protocol.NewAckTracker(
		protocol.TrackerCallbacks{
			Retry: func(message protocol.Message) error {
				return m.SendPacket(context.Background(), message)
			},

			Failed: func(message protocol.Message, err error) {
				log.Printf(
					"Device failed to acknowledge message %s: %v",
					message.ID,
					err,
				)

				m.responses.Fail(
					message.ID,
					fmt.Errorf("command delivery failed: %w", err),
				)
			},

			Acked: func(message protocol.Message, ack protocol.Acknowledgement) {
				if ack.Status != protocol.AckOk {
					m.responses.Fail(
						message.ID,
						fmt.Errorf(
							"device rejected command: %s: %s",
							ack.Status,
							ack.Error,
						),
					)
					return
				}

				log.Printf("Device acknowledged message %s", message.ID)
			},
		},
	)

	return m
}

func (m *Manager) Run(ctx context.Context) {
	go m.tracker.Run(ctx)
	m.responses.Run(ctx)
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

func (m *Manager) SendPacket(
	ctx context.Context,
	packet protocol.PacketVariant,
) error {
	data, err := protocol.SerialisePacket(packet)
	if err != nil {
		return err
	}

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
		data,
	)
}

func (m *Manager) SendTracked(
	ctx context.Context,
	message protocol.Message,
	policy protocol.RetryPolicy,
) error {
	if err := m.tracker.Track(message, policy); err != nil {
		return err
	}

	if err := m.SendPacket(ctx, message); err != nil {
		m.tracker.Cancel(message.ID)
		return err
	}

	return nil
}

func (m *Manager) SendAndWait(
	ctx context.Context,
	message protocol.Message,
	policy protocol.RetryPolicy,
	timeout time.Duration,
) (protocol.Message, error) {
	resultCH, err := m.responses.Track(message, timeout)
	if err != nil {
		return protocol.Message{}, err
	}
	defer m.responses.Cancel(message.ID)

	if err := m.SendTracked(ctx, message, policy); err != nil {
		return protocol.Message{}, err
	}

	select {
	case result := <-resultCH:
		return result.Message, result.Err

	case <-ctx.Done():
		return protocol.Message{}, ctx.Err()
	}
}
