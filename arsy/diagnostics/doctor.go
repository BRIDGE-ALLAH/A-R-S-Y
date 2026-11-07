package diagnostics

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/spf13/cobra"
)

// NewDoctorCommand creates the local A-R-S-Y diagnostics command.
func NewDoctorCommand(app core.App, version, secretKeyEnv string) *cobra.Command {
	command := &cobra.Command{
		Use:          "doctor",
		Short:        "Check local A-R-S-Y readiness",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, _ []string) error {
			status, err := inspect(app.DataDir(), version, secretKeyEnv, func() error {
				_, queryErr := app.ConcurrentDB().NewQuery("SELECT 1").Execute()
				return queryErr
			})
			fmt.Fprintf(command.OutOrStdout(), "A-R-S-Y doctor\n")
			fmt.Fprintf(command.OutOrStdout(), "Version: %s\n", status.Version)
			fmt.Fprintf(command.OutOrStdout(), "Go: %s (%s/%s)\n", status.GoVersion, status.GOOS, status.GOARCH)
			fmt.Fprintf(command.OutOrStdout(), "Data directory: %s\n", status.DataDirectory)
			fmt.Fprintf(command.OutOrStdout(), "Database: %s\n", status.Database)
			fmt.Fprintf(command.OutOrStdout(), "Secret key: %s\n", status.SecretKey)
			return err
		},
	}

	return command
}

type doctorStatus struct {
	Version       string
	GoVersion     string
	GOOS          string
	GOARCH        string
	DataDirectory string
	Database      string
	SecretKey     string
}

func inspect(dataDir, version, secretKeyEnv string, checkDatabase func() error) (doctorStatus, error) {
	status := doctorStatus{
		Version:       version,
		GoVersion:     runtime.Version(),
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
		DataDirectory: "accessible",
		Database:      "ready",
		SecretKey:     "not configured",
	}

	var failures []error
	if err := checkDataDirectory(dataDir); err != nil {
		status.DataDirectory = "unavailable"
		failures = append(failures, errors.New("local data directory is not accessible"))
	}
	if checkDatabase == nil || checkDatabase() != nil {
		status.Database = "not ready"
		failures = append(failures, errors.New("local database is not ready"))
	}
	if secretKeyEnv != "" {
		if value, ok := os.LookupEnv(secretKeyEnv); ok && value != "" {
			status.SecretKey = "configured"
		}
	}

	return status, errors.Join(failures...)
}

func checkDataDirectory(path string) error {
	file, err := os.CreateTemp(path, ".arsy-doctor-*")
	if err != nil {
		return err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	return os.Remove(name)
}
