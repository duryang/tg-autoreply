package config

import "github.com/BurntSushi/toml"

func LoadConfig(path string) (*Config, error) {
	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadSecrets(path string) (*Secrets, error) {
	var secrets Secrets

	if _, err := toml.DecodeFile(path, &secrets); err != nil {
		return nil, err
	}

	return &secrets, nil
}
