package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`

	// Scraper configuration
	BaseURL    string        `json:"base_url"`
	UserAgent  string        `json:"user_agent"`
	Timeout    time.Duration `json:"timeout"`
	RateLimit  time.Duration `json:"rate_limit"`
	MaxRetries int           `json:"max_retries"`

	// CLI configuration
	OutputFormat string `json:"output_format"`
	OutputFile   string `json:"output_file"`
	Verbose      bool   `json:"verbose"`

	// API configuration
	EnableCORS     bool     `json:"enable_cors"`
	AllowedOrigins []string `json:"allowed_origins"`

	// Cache configuration
	EnableCache bool          `json:"enable_cache"`
	CacheTTL    time.Duration `json:"cache_ttl"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:           "3030",
		Host:           "0.0.0.0",
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		BaseURL:        "https://hianime.to",
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Timeout:        30 * time.Second,
		RateLimit:      500 * time.Millisecond,
		MaxRetries:     3,
		OutputFormat:   "json",
		Verbose:        true,
		EnableCORS:     true,
		AllowedOrigins: []string{"*"},
		EnableCache:    true,
		CacheTTL:       5 * time.Minute,
	}
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	if port := os.Getenv("PORT"); port != "" {
		c.Port = port
	}

	if host := os.Getenv("HOST"); host != "" {
		c.Host = host
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		c.BaseURL = baseURL
	}

	if userAgent := os.Getenv("USER_AGENT"); userAgent != "" {
		c.UserAgent = userAgent
	}

	if timeoutStr := os.Getenv("TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			c.Timeout = timeout
		}
	}

	if rateLimitStr := os.Getenv("RATE_LIMIT"); rateLimitStr != "" {
		if rateLimit, err := time.ParseDuration(rateLimitStr); err == nil {
			c.RateLimit = rateLimit
		}
	}

	if maxRetriesStr := os.Getenv("MAX_RETRIES"); maxRetriesStr != "" {
		if maxRetries, err := strconv.Atoi(maxRetriesStr); err == nil {
			c.MaxRetries = maxRetries
		}
	}

	if outputFormat := os.Getenv("OUTPUT_FORMAT"); outputFormat != "" {
		c.OutputFormat = outputFormat
	}

	if verboseStr := os.Getenv("VERBOSE"); verboseStr != "" {
		if verbose, err := strconv.ParseBool(verboseStr); err == nil {
			c.Verbose = verbose
		}
	}

	if enableCORSStr := os.Getenv("ENABLE_CORS"); enableCORSStr != "" {
		if enableCORS, err := strconv.ParseBool(enableCORSStr); err == nil {
			c.EnableCORS = enableCORS
		}
	}

	if enableCacheStr := os.Getenv("ENABLE_CACHE"); enableCacheStr != "" {
		if enableCache, err := strconv.ParseBool(enableCacheStr); err == nil {
			c.EnableCache = enableCache
		}
	}

	if cacheTTLStr := os.Getenv("CACHE_TTL"); cacheTTLStr != "" {
		if cacheTTL, err := time.ParseDuration(cacheTTLStr); err == nil {
			c.CacheTTL = cacheTTL
		}
	}
}

// New creates a new configuration instance with defaults, env vars, and flags applied
func New() *Config {
	config := DefaultConfig()
	config.LoadFromEnv()
	return config
}
