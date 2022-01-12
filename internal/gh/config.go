package gh

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type GithubConfig struct {
	AccessToken string `json:"access_token"`
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, ".config", "ghcv-cli", "config.json")
	return path, nil
}

func LoadConfig() (*GithubConfig, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg GithubConfig
	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveConfig(cfg *GithubConfig) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0666)
}
