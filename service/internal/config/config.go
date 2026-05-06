package config

import (
	"log"
	"sync"
	"time"
)

type Config struct {
	Server struct {
		Host string
		Port int
	}

	Debug bool

	AdminBearerToken string

	Services []ServiceConfig

	EmailAccounts []EmailAccountConfig

	Sms struct {
		FortySixElks FortySixElksConfig
	}
}

type ServiceConfig struct {
	ID                     string
	Name                   string
	Owner                  string
	Scope                  string
	EmailAccessMode        string
	AllowedEmailAccountIDs []string
	Status                 string
	PublicKey              string
	Notes                  string
	CreatedAt              time.Time
	LastRerollAt           *time.Time
}

type EmailAccountConfig struct {
	ID           string
	Address      string
	DisplayName  string
	IsDefault    bool
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastTestedAt *time.Time
	SMTP         struct {
		Host     string
		Port     int
		Username string
		Password string
	}
}

type FortySixElksConfig struct {
	Username      string
	Password      string
	Status        string
	LastSyncedAt  time.Time
	RotationCount int
}

var (
	currentConfig *Config
	configMutex   sync.RWMutex
)

func Set(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	currentConfig = cfg
}

func Reset() {
	Set(nil)
}

func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}

func DebugLog(format string, v ...interface{}) {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if currentConfig != nil && currentConfig.Debug {
		log.Printf(format, v...)
	}
}
