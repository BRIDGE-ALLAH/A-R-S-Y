package config

import "testing"

func TestLoadFromDefaults(t *testing.T) {
	cfg, err := LoadFrom(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataDir != DefaultDataDir || cfg.APIPrefix != DefaultAPIPrefix || cfg.ConsolePath != DefaultConsolePath || cfg.SecretKeyEnv != DefaultSecretKeyEnv {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	values := map[string]string{
		"ARSY_DATA_DIR": "/var/lib/arsy",
		"ARSY_API_PREFIX": "/control/api/v1",
		"ARSY_CONSOLE_PATH": "/control",
		"ARSY_SECRET_KEY_ENV": "CONTROL_SECRET",
	}
	cfg, err := LoadFrom(func(key string) string { return values[key] })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataDir != values["ARSY_DATA_DIR"] || cfg.APIPrefix != values["ARSY_API_PREFIX"] || cfg.ConsolePath != values["ARSY_CONSOLE_PATH"] || cfg.SecretKeyEnv != values["ARSY_SECRET_KEY_ENV"] {
		t.Fatalf("environment values were not loaded: %#v", cfg)
	}
}

func TestLoadFromRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		key string
		value string
	}{
		{"relative API prefix", "ARSY_API_PREFIX", "api/v1"},
		{"traversal console path", "ARSY_CONSOLE_PATH", "/arsy/../admin"},
		{"invalid secret env name", "ARSY_SECRET_KEY_ENV", "secret-key"},
		{"overlapping routes", "ARSY_CONSOLE_PATH", "/api/arsy/v1/console"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFrom(func(key string) string {
				if key == tt.key { return tt.value }
				return ""
			})
			if err == nil { t.Fatal("expected invalid configuration to fail") }
		})
	}
}
