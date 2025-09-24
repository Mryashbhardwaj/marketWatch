package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Define Go structs matching the YAML structure
type Config struct {
	MutualFunds MutualFundConfig `yaml:"mutual_funds"`
	Equity      EquityConfig     `yaml:"equity"`
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
