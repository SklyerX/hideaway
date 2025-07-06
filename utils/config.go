package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	HashedPassword []byte `json:"hashed_password"`
	Salt           []byte `json:"salt"`
}

func ReadConfig() (Config, error) {
	filePaths := GetAppPaths()
	configFilePath := filepath.Join(filePaths["userData"], "config.json")
	data, err := os.ReadFile(configFilePath)

	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
