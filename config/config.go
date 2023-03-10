package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	LogLevel   string `yaml:"LogLevel"`
	RocketChat struct {
		UserId    string
		User      string `yaml:"User"`
		Password  string `yaml:"Password"`
		AuthToken string `yaml:"Authtoken"`
		HostName  string `yaml:"HostName"`
		SSL       bool   `yaml:"SSL"`
		Port      uint16 `yaml:"Port"`
	} `yaml:"RocketChat"`
	OpenAI struct {
		HostName           string `yaml:"HostName"`
		ApiToken           string `yaml:"ApiToken"`
		CompletionEndpoint string `yaml:"CompletionEndpoint"`
		ModerationEndpoint string `yaml:"ModerationEndpoint"`
		Model              string `yaml:"Model"`
		HistorySize        int    `yaml:"HistorySize"`
		HistoryMaxLength   int    `yaml:"HistoryMaxLength"`
		PrePrompt          string `yaml:"PrePrompt"`
		InputModeration    bool   `yaml:"InputModeration"`
		OutputModeration   bool   `yaml:"OutputModeration"`
	} `yaml:"OpenAI"`
}

func NewConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read configfile: %w", err)
	}

	// Parse YAML file
	var config Config

	// Default values
	config.RocketChat.SSL = true

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("cannot parse configfile %w", err)
	}

	return &config, nil
}
