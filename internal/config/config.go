package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MailServer struct {
	IMAPHost string `json:"imap_host"`
	IMAPPort string `json:"imap_port"`
	SMTPHost string `json:"smtp_host"`
	SMTPPort string `json:"smtp_port"`
}

type Config struct {
	Port          string     `json:"port"`
	BaseURL       string     `json:"base_url"`
	SessionSecret string     `json:"session_secret"`
	Mail          MailServer `json:"mail"`
}

func Load() (*Config, error) {
	path := "config.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Port == "" {
		c.Port = "8080"
	}
	if c.BaseURL == "" {
		c.BaseURL = "http://localhost:" + c.Port
	}
	if c.SessionSecret == "" {
		return fmt.Errorf("session_secret must be set in config")
	}
	if c.Mail.IMAPHost == "" {
		return fmt.Errorf("mail.imap_host must be set in config")
	}
	if c.Mail.IMAPPort == "" {
		c.Mail.IMAPPort = "993"
	}
	if c.Mail.SMTPHost == "" {
		return fmt.Errorf("mail.smtp_host must be set in config")
	}
	if c.Mail.SMTPPort == "" {
		c.Mail.SMTPPort = "587"
	}
	return nil
}
