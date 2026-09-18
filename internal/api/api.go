package api

import "context"

type DeviceCommander interface {
	SendCommand(ctx context.Context, command string) error
}

type API struct {
	devices DeviceCommander
	token   string
}

func New(devices DeviceCommander, token string) *API {
	return &API{
		devices: devices,
		token:   token,
	}
}
