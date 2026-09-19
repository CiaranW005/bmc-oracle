package main

import (
	"log"
	"net/http"

	"github.com/CiaranW005/bmc-oracle/internal/api"
	"github.com/CiaranW005/bmc-oracle/internal/auth"
	"github.com/CiaranW005/bmc-oracle/internal/config"
	"github.com/CiaranW005/bmc-oracle/internal/device"
	"github.com/CiaranW005/bmc-oracle/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	deviceAuth := device.NewBearerAuthenticator(cfg.DeviceToken)
	deviceManager := device.NewManager(deviceAuth)

	bmcAuth := auth.NewManager(cfg.PasswordHash, cfg.Enviroment == config.Production)
	bmcAPI := api.New(deviceManager, bmcAuth)

	mux := http.NewServeMux()

	mux.HandleFunc("/device", deviceManager.Handler)
	mux.HandleFunc("POST /api/login", bmcAPI.Login)

	mux.Handle(
		"POST /api/logout",
		bmcAuth.RequireAuth(
			http.HandlerFunc(bmcAPI.Logout),
		),
	)

	mux.Handle(
		"GET /api/session",
		bmcAuth.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	mux.Handle(
		"POST /api/wake",
		bmcAuth.RequireAuth(
			http.HandlerFunc(bmcAPI.Wake),
		),
	)

	log.Println("BMC Oracle listening on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", middleware.Logger(mux)))
}
