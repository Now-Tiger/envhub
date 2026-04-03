package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// ErrNoToken is returned when no authentication token is available
var ErrNoToken = errors.New("no token available")

// Config manages CLI configuration using Viper
type Config struct {
	mu         sync.RWMutex
	APIBaseURL string
	Token      string
	NoColor    bool
	Verbose    bool
	configPath string
}

var (
	// globalConfig is the singleton config instance
	globalConfig *Config
	once         sync.Once
)

// Init initializes the global configuration
func Init() *Config {
	once.Do(func() {
		globalConfig = &Config{
			APIBaseURL: "http://localhost:8080",
			NoColor:    false,
			Verbose:    false,
		}
	})
	return globalConfig
}

// Get returns the global config instance
func Get() *Config {
	once.Do(func() {
		globalConfig = &Config{
			APIBaseURL: "http://localhost:8080",
			NoColor:    false,
			Verbose:    false,
		}
	})
	return globalConfig
}

// Load loads configuration from file and environment
func (c *Config) Load(configPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set default values
	viper.SetDefault("api_base_url", "http://localhost:8080")

	// Handle config path
	if configPath != "" {
		c.configPath = configPath
		viper.SetConfigFile(configPath)
	} else {
		// Try to find config in standard locations
		home, err := homedir.Dir()
		if err != nil {
			return fmt.Errorf("failed to find home directory: %w", err)
		}

		configDir := filepath.Join(home, ".envhub")
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("ENVUB")
	viper.AutomaticEnv()

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Load values
	c.APIBaseURL = viper.GetString("api_base_url")
	c.NoColor = viper.GetBool("no_color")
	c.Verbose = viper.GetBool("verbose")

	// Load token (not cached in config file for security)
	c.Token = viper.GetString("token")
	if c.Token == "" {
		c.Token = os.Getenv("ENVUB_TOKEN")
	}

	return nil
}

// Token returns the stored authentication token
func (c *Config) TokenKey() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Token == "" {
		return "", ErrNoToken
	}
	return c.Token, nil
}

// SetToken stores the authentication token
func (c *Config) SetToken(token string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Token = token

	// Persist to config file
	if c.configPath != "" {
		viper.Set("token", token)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}
		// Set secure file permissions (0600) - owner read/write only
		if err := os.Chmod(c.configPath, 0600); err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}
	}

	return nil
}

// ClearToken removes the stored authentication token
func (c *Config) ClearToken() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Token = ""

	// Remove from config
	viper.Set("token", nil)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to clear token: %w", err)
	}

	return nil
}

// SetVerbose sets the verbose flag
func (c *Config) SetVerbose(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Verbose = v
}

// SetNoColor sets the no-color flag
func (c *Config) SetNoColor(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.NoColor = v
}
