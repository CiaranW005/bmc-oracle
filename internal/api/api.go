package api

import (
	"context"
	"net/http"
)

type DeviceCommander interface {
	SendCommand(ctx context.Context, command string) error
}

type Authenticator interface {
	CheckPassword(password string) bool
	StartSession(w http.ResponseWriter) error
	EndSession(w http.ResponseWriter, r *http.Request)
}

type API struct {
	devices DeviceCommander
	auth    Authenticator
}

func New(
	devices DeviceCommander,
	auth Authenticator,
) *API {
	return &API{
		devices: devices,
		auth:    auth,
	}
}
