package auth

import "net/http"

func (m *Manager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.cookieName)
		if err != nil {
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}

		if !m.ValidSession(cookie.Value) {
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
