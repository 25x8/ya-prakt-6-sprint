package config

import (
	"flag"
	"os"
)

// Config contains application configuration
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

// NewConfig creates a new configuration from environment variables or flags
func NewConfig() *Config {
	var cfg Config

	// Parse flags
	flag.StringVar(&cfg.RunAddress, "a", "", "Server run address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	// Override with env vars if present
	if envAddr := os.Getenv("RUN_ADDRESS"); envAddr != "" {
		cfg.RunAddress = envAddr
	}

	if envDBURI := os.Getenv("DATABASE_URI"); envDBURI != "" {
		cfg.DatabaseURI = envDBURI
	}

	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		cfg.AccrualSystemAddress = envAccrualAddr
	}

	// Set defaults if needed
	if cfg.RunAddress == "" {
		cfg.RunAddress = ":8080"
	}

	return &cfg
}
