package dhl

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	DHL24 DHL24Config `json:"dhl24"`
}

// DHL24Config contains DHL24 API credentials and settings
type DHL24Config struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	AccountNumber string `json:"accountNumber"`
	DebugFiles    bool   `json:"debugFiles"`
	DebugFilesDir string `json:"debugFilesDir"`
}

// LoadConfig reads configuration from config.json file
func LoadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w (copy config.example.json to config.json)", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	return &config, nil
}
