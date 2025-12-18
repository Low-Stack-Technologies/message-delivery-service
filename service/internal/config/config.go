package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type Config struct {
	mu sync.RWMutex

	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`

	Services []ServiceConfig `yaml:"services"`

	EmailAccounts []EmailAccountConfig `yaml:"email_accounts"`

	Sms struct {
		FortySixElks FortySixElksConfig `yaml:"46elks"`
	} `yaml:"sms"`
}

type ServiceConfig struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	PublicKey string `yaml:"public_key"`
}

type EmailAccountConfig struct {
	Address string `yaml:"address"`
	SMTP    struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"smtp"`
}

type FortySixElksConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var (
	currentConfig *Config
	configMutex   sync.RWMutex
)

const defaultConfig = `# Message Delivery Service Configuration

server:
  host: "0.0.0.0"
  port: 3000

# Service Authentication (Request Signing)
# Each service that uses this API needs a unique ID and its Ed25519 public key.
services:
  - id: "example-client"
    name: "Example Service"
    public_key: "base64_ed25519_public_key_here"

# Email SMTP accounts
# You can define multiple accounts. The "from" address in the request selects the account.
email_accounts:
  - address: "support@example.com"
    smtp:
      host: "smtp.example.com"
      port: 587
      username: "user@example.com"
      password: "password"

# SMS Providers
sms:
  46elks:
    username: "api_user_id"
    password: "api_password"
`

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Config file %s not found, creating default", path)
		if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	configMutex.Lock()
	currentConfig = &cfg
	configMutex.Unlock()

	return &cfg, nil
}

func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}

func Watch(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("Config file modified: %s, reloading...", event.Name)
					if _, err := Load(path); err != nil {
						log.Printf("Error reloading config: %v", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	return watcher.Add(path)
}
