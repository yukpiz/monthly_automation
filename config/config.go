package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	RecoruConfig *struct {
		LoginURL      string `yaml:"login_url"`
		HomeURL       string `yaml:"home_url"`
		ContractID    string `yaml:"contract_id"`
		LoginID       string `yaml:"login_id"`
		LoginPassword string `yaml:"login_password"`
	} `yaml:"recoru_config"`
}

func LoadConfig(fpath string) (*Config, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
