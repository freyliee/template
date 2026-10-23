package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	LogLevel        string
	ShutdownTimeout time.Duration

	DatabaseURL             string
	DatabaseMaxConns        int32
	DatabaseMinConns        int32
	DatabaseConnectTimeout  time.Duration
	DatabaseQueryTimeout    time.Duration
	DatabaseMaxConnLifetime time.Duration

	HTTPReadTimeout       time.Duration
	HTTPReadHeaderTimeout time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
}

func Load() (Config, error) {
	var cfg Config

	cfg.HTTPAddr = os.Getenv("HTTP_ADDR")
	if cfg.HTTPAddr == "" {
		return cfg, fmt.Errorf("HTTP_ADDR is required")
	}

	cfg.LogLevel = os.Getenv("LOG_LEVEL")
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return cfg, fmt.Errorf(
			"LOG_LEVEL must be debug, info, warn or error",
		)
	}

	value := os.Getenv("SHUTDOWN_TIMEOUT")
	timeout, err := time.ParseDuration(value)
	if err != nil || timeout <= 0 {
		return cfg, fmt.Errorf(
			"SHUTDOWN_TIMEOUT must be a positive duration, for example 10s",
		)
	}
	cfg.ShutdownTimeout = timeout

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}

	cfg.DatabaseMaxConns, err = readInt32("DATABASE_MAX_CONNS")
	if err != nil {
		return cfg, err
	}
	if cfg.DatabaseMaxConns <= 0 {
		return cfg, fmt.Errorf("DATABASE_MAX_CONNS must be positive")
	}

	cfg.DatabaseMinConns, err = readInt32("DATABASE_MIN_CONNS")
	if err != nil {
		return cfg, err
	}
	if cfg.DatabaseMinConns < 0 ||
		cfg.DatabaseMinConns > cfg.DatabaseMaxConns {
		return cfg, fmt.Errorf(
			"DATABASE_MIN_CONNS must be between 0 and DATABASE_MAX_CONNS",
		)
	}

	cfg.DatabaseConnectTimeout, err = readDuration("DATABASE_CONNECT_TIMEOUT")
	if err != nil {
		return cfg, err
	}

	cfg.DatabaseQueryTimeout, err = readDuration("DATABASE_QUERY_TIMEOUT")
	if err != nil {
		return cfg, err
	}

	cfg.DatabaseMaxConnLifetime, err = readDuration("DATABASE_MAX_CONN_LIFETIME")
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func readInt32(name string) (int32, error) {
	value, err := strconv.ParseInt(os.Getenv(name), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an int32: %w", name, err)
	}

	return int32(value), nil
}

func readDuration(name string) (time.Duration, error) {
	value, err := time.ParseDuration(os.Getenv(name))
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", name, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}

	return value, nil
}
