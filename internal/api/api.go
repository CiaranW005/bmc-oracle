package api

import (
	"context"
	"net/http"
	"time"

	"github.com/CiaranW005/bmc-oracle/internal/auth"
	"github.com/CiaranW005/bmc-oracle/internal/device"
	"github.com/CiaranW005/go-protocol"
)

type DeviceCommander interface {
	SendAndWait(
		ctx context.Context,
		message protocol.Message,
		policy protocol.RetryPolicy,
		timeout time.Duration,
	) (protocol.Message, error)
	SubscribeStatus() (<-chan device.Status, func())
	BeginWake(id protocol.ID) bool
	FinishWake(id protocol.ID, success bool, message string)
}

type Authenticator interface {
	CheckPassword(password string) bool
	StartSession(w http.ResponseWriter) error
	EndSession(w http.ResponseWriter, r *http.Request)
}

type API struct {
	devices    DeviceCommander
	auth       Authenticator
	googleAuth *auth.GoogleAuthenticator
}

func New(
	devices DeviceCommander,
	auth Authenticator,
	googleAuth *auth.GoogleAuthenticator,
) *API {
	return &API{
		devices:    devices,
		auth:       auth,
		googleAuth: googleAuth,
	}
}
