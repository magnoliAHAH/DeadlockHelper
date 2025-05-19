package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DeadlockPath string `json:"deadlock_path"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".deadlockhelper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

func SaveConfig(cfg Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

func LoadConfig() (Config, error) {
	var cfg Config
	path, err := getConfigPath()
	if err != nil {
		return cfg, err
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // файл не найден — вернуть пустой конфиг без ошибки
		}
		return cfg, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&cfg)
	return cfg, err
}
