package device

import (
	"time"

	"github.com/CiaranW005/go-protocol"
)

func (m *Manager) SubscribeStatus() (<-chan Status, func()) {
	// SubscribeStatus immediately supplies the current snapshot, then updates.
	// The caller must invoke the returned unsubscribe function when finished.
	updates := make(chan Status, 1)

	m.statusMu.Lock()
	m.subscribers[updates] = struct{}{}
	updates <- m.status
	m.statusMu.Unlock()

	unsubscribe := func() {
		m.statusMu.Lock()
		delete(m.subscribers, updates)
		m.statusMu.Unlock()
	}

	return updates, unsubscribe
}

// Caller must hold statusMu.
// Replace a buffered snapshot when necessary so slow subscribers receive
// the latest state without blocking publication.
func (m *Manager) broadcastStatusLocked() {
	for updates := range m.subscribers {
		select {
		case updates <- m.status:
		default:

			select {
			case <-updates:
			default:
			}

			updates <- m.status
		}
	}
}

func (m *Manager) publishStatus(bmc, pc bool) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	m.status.BMC = bmc
	m.status.PC = pc

	// Update LastSeen only when the BMC responds successfully.
	// Leave it unchanged when reporting the BMC offline.
	if bmc {
		m.status.LastSeen = time.Now()
	}

	m.broadcastStatusLocked()
}

func (m *Manager) BeginWake(id protocol.ID) bool {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	// Wake progresses from sending to booting after acknowledgement, then to
	// online or failed. Reject another wake while either active state is present.
	current := m.status.Wake
	if current != nil &&
		(current.State == "sending" || current.State == "booting") {
		return false
	}

	m.status.Wake = &WakeProgress{
		ID:    id,
		State: "sending",
	}
	m.broadcastStatusLocked()
	return true
}

func (m *Manager) acknowledgeWake(id protocol.ID) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	current := m.status.Wake
	if current == nil || current.ID != id || current.State != "sending" {
		return
	}

	m.status.Wake = &WakeProgress{
		ID:    id,
		State: "booting",
	}
	m.broadcastStatusLocked()
}

func (m *Manager) FinishWake(
	id protocol.ID,
	success bool,
	message string,
) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	// Ignore stale completions so an older request cannot overwrite the current
	// wake operation or change an already finished result.
	current := m.status.Wake
	if current == nil || current.ID != id ||
		(current.State != "sending" && current.State != "booting") {
		return
	}

	progress := &WakeProgress{
		ID:    id,
		State: "failed",
		Error: message,
	}

	if success {
		progress.State = "online"
		progress.Error = ""

		m.status.BMC = true
		m.status.PC = true
		m.status.LastSeen = time.Now()
	}

	m.status.Wake = progress
	m.broadcastStatusLocked()
}
