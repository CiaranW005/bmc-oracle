package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	EnvDeviceToken  = "ESP32_DEVICE_TOKEN"
	EnvWakeAPIToken = "WAKE_API_TOKEN"
)

type Config struct {
	DeviceToken  string
	WakeAPIToken string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No local .env file found; using process enviroment")
	}

	cfg := Config{}

	requiredVars := []struct {
		key  string
		dest *string
	}{
		{EnvDeviceToken, &cfg.DeviceToken},
		{EnvWakeAPIToken, &cfg.WakeAPIToken},
	}

	for _, env := range requiredVars {
		value, err := requiredEnv(env.key)
		if err != nil {
			return Config{}, err
		}

		*env.dest = value
	}

	return cfg, nil
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)

	if value == "" {
		return "", fmt.Errorf("%s is not set", key)
	}

	return value, nil
}
