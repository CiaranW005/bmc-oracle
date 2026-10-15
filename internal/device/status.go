package device

import "time"

func (m *Manager) SubscribeStatus() (<-chan Status, func()) {
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

func (m *Manager) publishStatus(bmc, pc bool) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	m.status.BMC = bmc
	m.status.PC = pc

	if bmc {
		m.status.LastSeen = time.Now()
	}

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
