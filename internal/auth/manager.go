package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	Expires time.Time
}

type Manager struct {
	passwordHash []byte

	cookieName   string
	secureCookie bool

	mu       sync.RWMutex
	sessions map[[32]byte]Session
}

func NewManager(passwordHash string, secureCookie bool) *Manager {
	cookieName := "bmc_sesssion"

	if secureCookie {
		cookieName = "__Host-bmc_session"
	}

	return &Manager{
		passwordHash: []byte(passwordHash),

		cookieName:   cookieName,
		secureCookie: secureCookie,

		sessions: make(map[[32]byte]Session),
	}
}

func (m *Manager) CheckPassword(password string) bool {
	log.Printf("Password hash length: %d", len(m.passwordHash))

	return bcrypt.CompareHashAndPassword(
		m.passwordHash,
		[]byte(password),
	) == nil
}

func (m *Manager) StartSession(w http.ResponseWriter) error {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	key := sha256.Sum256([]byte(token))

	m.mu.Lock()
	m.sessions[key] = Session{
		Expires: time.Now().Add(8 * time.Hour),
	}
	m.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   8 * 60 * 60,
	})

	return nil
}

func (m *Manager) ValidSession(token string) bool {
	key := sha256.Sum256([]byte(token))

	m.mu.RLock()
	session, exists := m.sessions[key]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(session.Expires) {
		m.mu.Lock()
		delete(m.sessions, key)
		m.mu.Unlock()

		return false
	}

	return true
}

func (m *Manager) EndSession(
	w http.ResponseWriter,
	r *http.Request,
) {
	cookie, err := r.Cookie(m.cookieName)

	if err == nil {
		key := sha256.Sum256([]byte(cookie.Value))

		m.mu.Lock()
		delete(m.sessions, key)
		m.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
