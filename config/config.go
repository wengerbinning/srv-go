package config

import (
	"os"
	"fmt"
	"gopkg.in/yaml.v3"
)

type Conf struct {
	Storage struct {
		SrvPath    string `yaml:"srv_path"`
		SrvDbName  string `yaml:"srv_db_name"`
		UsrPath    string `yaml:"usr_path"`
	} `yaml:"storage"`

	Log struct {
		Stdio       bool `yaml:"stdio"`
		File        bool `yaml:"file"`
		FilePath  string `yaml:"file_path"`
		Syslog      bool `yaml:"syslog"`
		Systag    string `yaml:"systag"`
	} `yaml:"log"`

	User struct {
		Root      string `yaml:"root"`
		Default   string `yaml:"default"`
	}

	Network struct {
		Http struct {
			Enable       bool   `yaml:"enable"`
			Listen       string `yaml:"listen"`
			Port         string `yaml:"port"`
			ReadTimeout  string `yaml:"read_timeout"`
			WriteTimeout string `yaml:"write_timeout"`
		} `yaml:"http"`
	} `yaml:"network"`
}

func Load(path string) (*Conf, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil
	}

	conf := Default()
	if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return conf, nil
}
