package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabasePath       string
	BaseHostname       string
	TokenTTL           time.Duration
	ShutdownTimeout    time.Duration
	SonyflakeMachineID int
}

func Load() (Config, error) {
	tokenTTL, err := durationEnv("AUTH_TOKEN_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	machineID, err := intEnv("SONYFLAKE_MACHINE_ID", 0)
	if err != nil {
		return Config{}, err
	}
	if machineID < 0 || machineID > 65535 {
		return Config{}, fmt.Errorf("SONYFLAKE_MACHINE_ID must be between 0 and 65535")
	}

	baseHostname := strings.ToLower(env("DULIO_BASE_HOSTNAME", "3dreamstudio.com.br"))
	if !validHostname(baseHostname) {
		return Config{}, fmt.Errorf("DULIO_BASE_HOSTNAME must be a valid hostname without a scheme, port, or path")
	}

	return Config{
		DatabasePath:       env("DATABASE_PATH", "./data/dulio.db"),
		BaseHostname:       baseHostname,
		TokenTTL:           tokenTTL,
		ShutdownTimeout:    shutdownTimeout,
		SonyflakeMachineID: machineID,
	}, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return value, nil
}

func intEnv(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return value, nil
}

func validHostname(value string) bool {
	if value == "" || len(value) > 253 || strings.HasSuffix(value, ".") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') && character != '-' {
				return false
			}
		}
	}
	return true
}
