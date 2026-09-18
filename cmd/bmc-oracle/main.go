package main

import (
	"log"
	"net/http"

	"github.com/CiaranW005/bmc-oracle/internal/api"
	"github.com/CiaranW005/bmc-oracle/internal/config"
	"github.com/CiaranW005/bmc-oracle/internal/device"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	deviceAuth := device.NewBearerAuthenticator(cfg.DeviceToken)
	deviceManager := device.NewManager(deviceAuth)

	bmcAPI := api.New(deviceManager, cfg.WakeAPIToken)

	mux := http.NewServeMux()

	mux.HandleFunc("/device", deviceManager.Handler)
	http.HandleFunc("/wake", bmcAPI.Wake)

	log.Println("BMC Oracle listening on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
