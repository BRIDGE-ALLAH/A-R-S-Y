package config

import "testing"

func TestDefault(t *testing.T) {
	got := Default()
	want := Config{
		DataDir:      "",
		APIPrefix:    "/api/arsy/v1",
		ConsoleRoute: "/arsy",
		SecretKeyEnv: "ARSY_SECRET_KEY",
	}
	if got != want {
		t.Fatalf("Default() = %#v, want %#v", got, want)
	}
}

func TestLoadFromEnv(t *testing.T) {
	values := map[string]string{
		"ARSY_DATA_DIR":       "./control-data",
		"ARSY_API_PREFIX":     "/control/api/v1",
		"ARSY_CONSOLE_ROUTE":  "/control",
		"ARSY_SECRET_KEY_ENV": "CONTROL_PLANE_KEY",
	}
	got, err := LoadFromEnv(func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	})
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if got.DataDir != values["ARSY_DATA_DIR"] || got.APIPrefix != values["ARSY_API_PREFIX"] ||
		got.ConsoleRoute != values["ARSY_CONSOLE_ROUTE"] || got.SecretKeyEnv != values["ARSY_SECRET_KEY_ENV"] {
		t.Fatalf("LoadFromEnv() = %#v", got)
	}
}

func TestLoadFromEnvRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
	}{
		{name: "empty data directory", values: map[string]string{"ARSY_DATA_DIR": " "}},
		{name: "relative API prefix", values: map[string]string{"ARSY_API_PREFIX": "api/v1"}},
		{name: "query in console route", values: map[string]string{"ARSY_CONSOLE_ROUTE": "/arsy?debug=1"}},
		{name: "invalid secret key env name", values: map[string]string{"ARSY_SECRET_KEY_ENV": "secret-key"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := LoadFromEnv(func(name string) (string, bool) {
				value, ok := test.values[name]
				return value, ok
			})
			if err == nil {
				t.Fatal("expected invalid configuration to fail")
			}
		})
	}
}
