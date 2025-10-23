package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	DSN               string `mapstructure:"dsn"`                              // data source name e.g.: postgres://user:password@host:123/database?sslmode=disable
	MinOpenConnection int    `mapstructure:"min_open_connection" default:"5"`  // minimum open DB connections
	MaxOpenConnection int    `mapstructure:"max_open_connection" default:"20"` // maximum allowed open DB connections
}

type AuthConfig struct {
	SecretKey                  string `yaml:"secret_key"`
	TokenInactivityMinutes     int    `yaml:"token_inactivity_minutes"`
	TokenAbsoluteExpiryMinutes int    `yaml:"token_absolute_expiry_minutes"`
}

// Define Go structs matching the YAML structure
type Config struct {
	MutualFunds MutualFundConfig `yaml:"mutual_funds"`
	Equity      EquityConfig     `yaml:"equity"`
	DB          DBConfig         `yaml:"db"`
	AuthConfig  AuthConfig       `yaml:"auth"`
}

type MutualFundConfig struct {
	TradeFilesDirectory string `yaml:"tradefiles_diretory"`
}

type EquityConfig struct {
	TradeFilesDirectory string `yaml:"tradefiles_diretory"`
}

// LoadConfig reads and parses the YAML config file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	fmt.Println("------config---------------------------")
	fmt.Println(string(data))
	fmt.Println("---------------------------------------")
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
