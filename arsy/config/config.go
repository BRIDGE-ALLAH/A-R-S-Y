package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var environmentNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Config contains the A-R-S-Y control-plane runtime settings.
type Config struct {
	DataDir      string
	APIPrefix    string
	ConsoleRoute string
	SecretKeyEnv string
}

// Default returns the single source of defaults for A-R-S-Y configuration.
func Default() Config {
	return Config{
		DataDir:      "",
		APIPrefix:    "/api/arsy/v1",
		ConsoleRoute: "/arsy",
		SecretKeyEnv: "ARSY_SECRET_KEY",
	}
}

// Load reads and validates the A-R-S-Y-specific environment settings.
func Load() (Config, error) {
	return LoadFromEnv(os.LookupEnv)
}

// LoadFromEnv loads settings through lookup, which makes environment behavior testable.
func LoadFromEnv(lookup func(string) (string, bool)) (Config, error) {
	if lookup == nil {
		return Config{}, fmt.Errorf("environment lookup must not be nil")
	}

	config := Default()
	if value, ok := lookup("ARSY_DATA_DIR"); ok {
		config.DataDir = value
	}
	if value, ok := lookup("ARSY_API_PREFIX"); ok {
		config.APIPrefix = value
	}
	if value, ok := lookup("ARSY_CONSOLE_ROUTE"); ok {
		config.ConsoleRoute = value
	}
	if value, ok := lookup("ARSY_SECRET_KEY_ENV"); ok {
		config.SecretKeyEnv = value
	}

	if strings.TrimSpace(config.DataDir) == "" || strings.ContainsRune(config.DataDir, '\x00') {
		return Config{}, fmt.Errorf("ARSY_DATA_DIR must be a non-empty filesystem path")
	}
	if err := validateRoute("ARSY_API_PREFIX", config.APIPrefix); err != nil {
		return Config{}, err
	}
	if err := validateRoute("ARSY_CONSOLE_ROUTE", config.ConsoleRoute); err != nil {
		return Config{}, err
	}
	if !environmentNamePattern.MatchString(config.SecretKeyEnv) {
		return Config{}, fmt.Errorf("ARSY_SECRET_KEY_ENV must name a valid environment variable")
	}

	return config, nil
}

func validateRoute(name, value string) error {
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") ||
		strings.HasSuffix(value, "/") || strings.ContainsAny(value, "?#\\\x00") ||
		strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must be an absolute URL path without query, fragment, or trailing slash", name)
	}
	return nil
}
