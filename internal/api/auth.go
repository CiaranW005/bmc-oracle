package api

import (
	"crypto/subtle"
	"net/http"
)

// To be changed later only mimics device token but will need to change for when different deivces need to connect
func (a *API) authenticated(r *http.Request) bool {
	expected := "Bearer " + a.token
	actual := r.Header.Get("Authorization")

	return subtle.ConstantTimeCompare(
		[]byte(actual),
		[]byte(expected),
	) == 1
}
