package config

import (
	"os"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Storage struct {
		SrvPath    string `yaml:"srv_path"`
		SrvDbName  string `yaml:"srv_db_name"`
		UsrPath    string `yaml:"usr_path"`
	} `yaml:"storage"`
}

func Default() *Config {
	cfg := &Config{}
	cfg.Storage.SrvPath = "./etc/srv"
	cfg.Storage.SrvDbName = "service.db"
	cfg.Storage.UsrPath = "./etc/usr"
	return cfg
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}
