package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Log        LogConfig        `mapstructure:"log"`
	Auth       AuthConfig       `mapstructure:"auth"`
	RateLimit  RateLimitConfig  `mapstructure:"ratelimit"`
	Moderation ModerationConfig `mapstructure:"moderation"`
	Matchers   MatchersConfig   `mapstructure:"matchers"`
}

// MatchersConfig holds matcher configurations
type MatchersConfig struct {
	AC       ACConfig       `mapstructure:"ac"`
	Regex    RegexConfig    `mapstructure:"regex"`
	External ExternalConfig `mapstructure:"external"`
}

// ACConfig is the configuration for AC automaton matcher
type ACConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// RegexConfig is the configuration for regex matcher
type RegexConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ExternalConfig is the configuration for external API matcher
type ExternalConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	APIURL  string        `mapstructure:"api_url"`
	APIKey  string        `mapstructure:"api_key"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	APIKeys []string `mapstructure:"api_keys"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	Rate     float64 `mapstructure:"rate"`
	Capacity int     `mapstructure:"capacity"`
}

// LogConfig holds logging settings
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

// ModerationConfig holds moderation engine settings
type ModerationConfig struct {
	PipelineMode      string  `mapstructure:"pipeline_mode"`
	WeightedThreshold float64 `mapstructure:"weighted_threshold"`
	FallbackAction    string  `mapstructure:"fallback_action"`
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
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")

	// Auth defaults
	v.SetDefault("auth.api_keys", []string{})

	// Rate limit defaults
	v.SetDefault("ratelimit.rate", 100.0)
	v.SetDefault("ratelimit.capacity", 200)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "json")

	// Moderation defaults
	v.SetDefault("moderation.pipeline_mode", "chain")
	v.SetDefault("moderation.weighted_threshold", 0.5)
	v.SetDefault("moderation.fallback_action", "pass")

	// Matcher defaults
	v.SetDefault("matchers.ac.enabled", true)
	v.SetDefault("matchers.regex.enabled", false)
	v.SetDefault("matchers.external.enabled", false)
	v.SetDefault("matchers.external.api_url", "")
	v.SetDefault("matchers.external.api_key", "")
	v.SetDefault("matchers.external.timeout", "5s")
}
