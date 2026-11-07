package deployments

import (
	"context"
	"strings"
	"testing"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestParseComposePSMapsStates(t *testing.T) {
	cases := []struct {
		name    string
		output  string
		want    []ContainerStatus
		wantErr bool
	}{
		{
			name:   "array running with publishers",
			output: `[{"Service":"web","Image":"nginx:1","State":"running","Health":"healthy","Publishers":[{"URL":"0.0.0.0","TargetPort":80,"PublishedPort":8080}]}]`,
			want:   []ContainerStatus{{Service: "web", Image: "nginx:1", State: "running", Health: "healthy", Ports: "0.0.0.0:8080->80"}},
		},
		{
			name:   "ndjson exited with legacy ports",
			output: "{\"Service\":\"api\",\"Image\":\"api:2\",\"State\":\"exited\",\"Ports\":\"0.0.0.0:9090->9090/tcp\"}",
			want:   []ContainerStatus{{Service: "api", Image: "api:2", State: "exited", Ports: "0.0.0.0:9090->9090/tcp"}},
		},
		{
			name:   "restarting service",
			output: `[{"Service":"worker","State":"restarting"}]`,
			want:   []ContainerStatus{{Service: "worker", State: "restarting"}},
		},
		{
			name:   "unhealthy overrides running",
			output: `[{"Service":"web","State":"running","Health":"unhealthy"}]`,
			want:   []ContainerStatus{{Service: "web", State: "unhealthy", Health: "unhealthy"}},
		},
		{
			name:   "missing empty container state",
			output: `[{"Service":"cron","State":""}]`,
			want:   []ContainerStatus{{Service: "cron", State: "missing"}},
		},
		{
			name:   "unknown state passes through",
			output: `[{"Service":"cache","State":"paused"}]`,
			want:   []ContainerStatus{{Service: "cache", State: "paused"}},
		},
		{name: "invalid json", output: "{bad", wantErr: true},
		{name: "empty output", output: "", wantErr: true},
		{name: "empty array", output: "[]", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseComposePS(tc.output)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseComposePS() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if len(got) != len(tc.want) {
				t.Fatalf("ParseComposePS() returned %d services, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("ParseComposePS()[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestMissingServicesReportsAbsentManifestServices(t *testing.T) {
	manifest, err := ValidateCompose("services:\n  api:\n    image: api\n  web:\n    image: nginx\n")
	if err != nil {
		t.Fatal(err)
	}
	services := []ContainerStatus{{Service: "web", State: "running"}}
	got := MissingServices(manifest, services)
	if len(got) != 1 || got[0] != "api" {
		t.Fatalf("MissingServices() = %v, want [api]", got)
	}
	if len(MissingServices(manifest, []ContainerStatus{{Service: "api"}, {Service: "web"}})) != 0 {
		t.Fatal("MissingServices() reported services that were present")
	}
}

func TestInspectDeploymentStatusValidatesInput(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	if _, err := InspectDeploymentStatus(context.Background(), app, "short", "ARSY_SECRET_KEY"); err == nil || !strings.Contains(err.Error(), "deployment ID is invalid") {
		t.Fatalf("InspectDeploymentStatus() error = %v, want invalid deployment ID", err)
	}
}

func TestInspectDeploymentStatusRejectsUnknownDeployment(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deploymentID, err := NewDeploymentID()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := InspectDeploymentStatus(context.Background(), app, deploymentID, "ARSY_SECRET_KEY"); err == nil || !strings.Contains(err.Error(), "deployment not found") {
		t.Fatalf("InspectDeploymentStatus() error = %v, want deployment not found", err)
	}
}

func TestInspectDeploymentStatusRejectsMissingCredentials(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "inspect-a", StateSucceeded)
	application, err := app.FindRecordById("arsy_applications", deployment.GetString("application"))
	if err != nil {
		t.Fatal(err)
	}
	server, err := app.FindRecordById("arsy_servers", application.GetString("server"))
	if err != nil {
		t.Fatal(err)
	}
	server.Set("host_key_fingerprint", "SHA256:inspect-test-fingerprint")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}
	_, err = InspectDeploymentStatus(context.Background(), app, deployment.GetString("deployment_id"), "ARSY_SECRET_KEY")
	if err == nil || !strings.Contains(err.Error(), "stored SSH credential could not be decrypted") {
		t.Fatalf("InspectDeploymentStatus() error = %v, want credential decryption failure", err)
	}
}
