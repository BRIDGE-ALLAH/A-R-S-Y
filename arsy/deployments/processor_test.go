package deployments

import "testing"

func TestAllComposeContainersRunning(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{"all running array", `[{"State":"running"},{"State":"running"}]`, false},
		{"all running lines", "{\"State\":\"running\"}\n{\"State\":\"running\"}", false},
		{"exited service", `[{"State":"running"},{"State":"exited"}]`, true},
		{"empty", "[]", true},
		{"invalid JSON", "{bad", true},
		{"empty output", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AllComposeContainersRunning(tt.output); (err != nil) != tt.wantErr {
				t.Fatalf("AllComposeContainersRunning() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
