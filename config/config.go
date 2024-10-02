package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

var (
	config Config
	once   sync.Once
)

type Config struct {
	Telegram struct {
		Token string `json:"token"`
	} `json:"telegram"`
	Git struct {
		Owner      string `json:"owner"`
		Repo       string `json:"repo"`
		Token      string `json:"token"`
		WorkflowID string `json:"workflowID"`
		Branch     string `json:"branch"`
	} `json:"git"`
}

func Get() *Config {
	once.Do(func() {
		file, err := os.Open("init")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
		if err := json.NewDecoder(file).Decode(&config); err != nil {
			log.Fatalf("Failed to decode config JSON: %v", err)
		}

		_, _ = json.MarshalIndent(config, "", "  ")
		//log.Printf("\nConfiguration:\n%v", string(jsonData))
	})
	return &config
}
