package servers

import "testing"

func TestParseLinuxFacts(t *testing.T) {
	tests := []struct {
		output string
		os     string
		arch   string
		wantErr bool
	}{
		{"Linux\nx86_64", "linux", "amd64", false},
		{"Linux\naarch64", "linux", "arm64", false},
		{"Darwin\nx86_64", "", "", true},
		{"Linux", "", "", true},
		{"Linux\nmips", "", "", true},
	}
	for _, tt := range tests {
		osName, arch, err := ParseLinuxFacts(tt.output)
		if (err != nil) != tt.wantErr || osName != tt.os || arch != tt.arch {
			t.Fatalf("ParseLinuxFacts(%q) = (%q, %q, %v)", tt.output, osName, arch, err)
		}
	}
}

func TestDockerReadinessDistinguishesFailureModes(t *testing.T) {
	tests := []struct {
		name string
		result CommandResult
		code string
		ready bool
	}{
		{"ready", CommandResult{Stdout: "27.1.1"}, "ready", true},
		{"missing", CommandResult{ExitCode: 127, Stderr: "docker: command not found"}, "docker_missing", false},
		{"daemon", CommandResult{ExitCode: 1, Stderr: "Cannot connect to the Docker daemon"}, "daemon_unavailable", false},
		{"permission", CommandResult{ExitCode: 1, Stderr: "permission denied while trying to connect"}, "permission_denied", false},
	}
	for _, tt := range tests {
		got := DockerReadiness(tt.result)
		if got.Code != tt.code || got.Ready != tt.ready {
			t.Fatalf("%s: readiness = %#v", tt.name, got)
		}
	}
}

func TestComposeReadiness(t *testing.T) {
	ready := ComposeReadiness(CommandResult{Stdout: "2.29.1"})
	if !ready.Ready || ready.Version != "2.29.1" {
		t.Fatalf("compose ready = %#v", ready)
	}
	missing := ComposeReadiness(CommandResult{ExitCode: 1, Stderr: "docker: 'compose' is not a docker command"})
	if missing.Ready || missing.Code != "compose_missing" {
		t.Fatalf("compose missing = %#v", missing)
	}
}

func TestBuildReadinessSummaryRequiresEveryComponent(t *testing.T) {
	ready := ComponentStatus{Ready: true, Code: "ready"}
	summary := BuildReadinessSummary(ready, ready, ready, ready, ready)
	if !summary.Ready {
		t.Fatal("all-ready components must produce ready summary")
	}
	summary = BuildReadinessSummary(ready, ready, ready, ComponentStatus{Code: "permission_denied"}, ready)
	if summary.Ready {
		t.Fatal("one failed component must produce not-ready summary")
	}
}
