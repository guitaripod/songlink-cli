package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config stores Apple Music API credentials and authentication details.
type Config struct {
	TeamID       string `json:"team_id"`       // Apple Developer Team ID
	KeyID        string `json:"key_id"`        // MusicKit API Key ID
	PrivateKey   string `json:"private_key"`   // P8 private key content
	MusicID      string `json:"music_id"`      // Music identifier (usually same as TeamID)
	ConfigExists bool   `json:"-"`             // Whether config file exists on disk
}

// GetConfigPath returns the path to the configuration file, creating the directory if needed.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".songlink-cli")
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig loads the configuration from disk, returning an empty config if the file doesn't exist.
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{ConfigExists: false}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.ConfigExists = true
	return &config, nil
}

// SaveConfig writes the configuration to disk with proper permissions.
func (c *Config) SaveConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}