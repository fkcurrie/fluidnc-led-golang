package config

import (
	"encoding/json"
	"os"

	"github.com/fkcurrie/fluidnc-led-golang/internal/types"
)

// Config represents the application configuration
type Config struct {
	Display types.DisplayConfig `json:"display"`
	GRBL    types.FluidNCConfig `json:"grbl"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Display: types.DisplayConfig{
			Width:      32,
			Height:     8,
			Brightness: 64,
		},
		GRBL: types.FluidNCConfig{
			Host: "localhost",
			Port: 23,
		},
	}
} 