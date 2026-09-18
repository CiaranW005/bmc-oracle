package device

import (
	"crypto/subtle"
	"net/http"
)

type Authenticator interface {
	Valid(r *http.Request) bool
}

type BearerAuthenticator struct {
	token string
}

func NewBearerAuthenticator(token string) *BearerAuthenticator {
	return &BearerAuthenticator{
		token: token,
	}
}

func (a *BearerAuthenticator) Valid(r *http.Request) bool {
	expected := "Bearer " + a.token
	actual := r.Header.Get("Authorization")

	return subtle.ConstantTimeCompare(
		[]byte(actual),
		[]byte(expected),
	) == 1
}
