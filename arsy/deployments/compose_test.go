package deployments

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateComposeAcceptsSupportedManifest(t *testing.T) {
	input := `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
    environment:
      APP_MODE: production
      FEATURE_FLAG: enabled
    volumes:
      - data:/var/lib/app
    networks:
      - frontend
    restart: unless-stopped
volumes:
  data: {}
networks:
  frontend: {}
`
	got, err := ValidateCompose(input)
	if err != nil {
		t.Fatal(err)
	}
	web, ok := got.Services["web"]
	if !ok || web.Image != "nginx:1.27" || len(web.PublishedPorts) != 1 || len(web.EnvironmentKeys) != 2 {
		t.Fatalf("unexpected validated service: %#v", web)
	}
	if _, ok := got.Volumes["data"]; !ok {
		t.Fatal("declared named volume was not retained")
	}
}

func TestValidateComposeRejectsUnsafeAndUnsupportedFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
		field string
	}{
		{"build", "services:\n  web:\n    build: .\n", "services.web.build"},
		{"privileged", "services:\n  web:\n    image: nginx\n    privileged: true\n", "services.web.privileged"},
		{"host network", "services:\n  web:\n    image: nginx\n    network_mode: host\n", "services.web.network_mode"},
		{"host pid", "services:\n  web:\n    image: nginx\n    pid: host\n", "services.web.pid"},
		{"device", "services:\n  web:\n    image: nginx\n    devices: []\n", "services.web.devices"},
		{"docker socket", "services:\n  web:\n    image: nginx\n    volumes:\n      - data:/var/run/docker.sock\nvolumes:\n  data: {}\n", "services.web.volumes[0]"},
		{"host path", "services:\n  web:\n    image: nginx\n    volumes:\n      - /host/path:/data\n", "services.web.volumes[0]"},
		{"missing image", "services:\n  web: {}\n", "services.web.image"},
		{"invalid port", "services:\n  web:\n    image: nginx\n    ports:\n      - '70000:80'\n", "services.web.ports[0]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateCompose(tt.input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || validationErr.Field != tt.field {
				t.Fatalf("error = %v, want validation field %q", err, tt.field)
			}
		})
	}
}

func TestValidateComposeRejectsMalformedYAMLAndEmptyServices(t *testing.T) {
	if _, err := ValidateCompose("services: ["); err == nil {
		t.Fatal("malformed YAML was accepted")
	}
	_, err := ValidateCompose("services: {}")
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) || validationErr.Field != "services" {
		t.Fatalf("empty services error = %v", err)
	}
}

func TestValidationErrorContainsFieldAndDoesNotEchoEnvironmentValue(t *testing.T) {
	_, err := ValidateCompose("services:\n  web:\n    image: nginx\n    environment:\n      BAD KEY: private-value\n")
	if err == nil || !strings.Contains(err.Error(), "services.web.environment.BAD KEY") || strings.Contains(err.Error(), "private-value") {
		t.Fatalf("unexpected environment validation error: %v", err)
	}
}
