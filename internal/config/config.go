package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Log        LogConfig        `mapstructure:"log"`
	Moderation ModerationConfig `mapstructure:"moderation"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// LogConfig holds logging settings
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

// ModerationConfig holds moderation engine settings
type ModerationConfig struct {
	PipelineMode   string `mapstructure:"pipeline_mode"`
	FallbackAction string `mapstructure:"fallback_action"`
}

// Addr returns the server address in host:port format
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Load reads configuration from YAML file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config file
	v.SetConfigName("default")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")
	v.AddConfigPath(".")

	// Enable environment variable override
	v.SetEnvPrefix("MODERATION")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults + env
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "json")

	// Moderation defaults
	v.SetDefault("moderation.pipeline_mode", "chain")
	v.SetDefault("moderation.fallback_action", "pass")
}
