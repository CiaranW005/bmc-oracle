package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Enviroment string

const (
	EnvDeviceToken  = "ESP32_DEVICE_TOKEN"
	EnvPasswordHash = "BMC_PASSWORD_HASH"
	EnvEnviroment   = "BMC_ENV"

	Development Enviroment = "development"
	Production  Enviroment = "production"
)

type Config struct {
	Enviroment   Enviroment
	DeviceToken  string
	PasswordHash string
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
		{EnvPasswordHash, &cfg.PasswordHash},
	}

	for _, env := range requiredVars {
		value, err := requiredEnv(env.key)
		if err != nil {
			return Config{}, err
		}

		*env.dest = value
	}

	enviroment, err := loadEnviroment()
	if err != nil {
		return Config{}, err
	}

	cfg.Enviroment = enviroment

	return cfg, nil
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)

	if value == "" {
		return "", fmt.Errorf("%s is not set", key)
	}

	return value, nil
}

func loadEnviroment() (Enviroment, error) {
	value, err := requiredEnv(EnvEnviroment)
	if err != nil {
		return "", err
	}

	env := Enviroment(value)

	switch env {
	case Development, Production:
		return env, nil
	default:
		return "", fmt.Errorf("Invalid BMC_ENV: %s", value)
	}
}
