package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultModel    = "llama3.2"
	DefaultProvider = "siliconflow"
	ConfigDir       = ".config/auto-git"
	ConfigFile      = "config.yaml"
)

type Config struct {
	Provider string `yaml:"provider"`
	Endpoint string `yaml:"endpoint"`
	Model    string `yaml:"model"`
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ConfigDir, ConfigFile), nil
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ConfigDir), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Provider: DefaultProvider,
			Model:    DefaultModel,
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for backward compatibility
	if config.Provider == "" {
		config.Provider = DefaultProvider
	}
	if config.Model == "" {
		config.Model = DefaultModel
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func SetModel(model string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.Model = model
	return SaveConfig(config)
}

func SetProvider(provider string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.Provider = provider
	return SaveConfig(config)
}

func SetEndpoint(endpoint string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.Endpoint = endpoint
	return SaveConfig(config)
}

