package config

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

// Config contains settings owned by the A-R-S-Y control plane.
type Config struct {
	DataDir      string
	APIPrefix    string
	ConsolePath  string
	SecretKeyEnv string
}

const (
	DefaultDataDir      = "pb_data/arsy"
	DefaultAPIPrefix    = "/api/arsy/v1"
	DefaultConsolePath  = "/arsy"
	DefaultSecretKeyEnv = "ARSY_SECRET_KEY"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Load reads A-R-S-Y settings from the process environment.
func Load() (Config, error) {
	return LoadFrom(os.Getenv)
}

// LoadFrom loads configuration using getenv. It is exported to make
// environment parsing deterministic in tests and embedding applications.
func LoadFrom(getenv func(string) string) (Config, error) {
	cfg := Config{
		DataDir:      valueOrDefault(getenv("ARSY_DATA_DIR"), DefaultDataDir),
		APIPrefix:    valueOrDefault(getenv("ARSY_API_PREFIX"), DefaultAPIPrefix),
		ConsolePath:  valueOrDefault(getenv("ARSY_CONSOLE_PATH"), DefaultConsolePath),
		SecretKeyEnv: valueOrDefault(getenv("ARSY_SECRET_KEY_ENV"), DefaultSecretKeyEnv),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

// Validate rejects malformed paths and environment variable names.
func (c Config) Validate() error {
	if c.DataDir == "" || strings.TrimSpace(c.DataDir) != c.DataDir {
		return fmt.Errorf("ARSY_DATA_DIR must be a non-empty path without surrounding whitespace")
	}
	for label, value := range map[string]string{
		"ARSY_API_PREFIX": c.APIPrefix,
		"ARSY_CONSOLE_PATH": c.ConsolePath,
	} {
		if !strings.HasPrefix(value, "/") || value == "/" || path.Clean(value) != value || strings.Contains(value, "//") {
			return fmt.Errorf("%s must be a clean absolute URL path", label)
		}
	}
	if c.APIPrefix == c.ConsolePath || strings.HasPrefix(c.ConsolePath+"/", c.APIPrefix+"/") || strings.HasPrefix(c.APIPrefix+"/", c.ConsolePath+"/") {
		return fmt.Errorf("ARSY_API_PREFIX and ARSY_CONSOLE_PATH must not overlap")
	}
	if !envNamePattern.MatchString(c.SecretKeyEnv) {
		return fmt.Errorf("ARSY_SECRET_KEY_ENV must be a valid environment variable name")
	}
	return nil
}
