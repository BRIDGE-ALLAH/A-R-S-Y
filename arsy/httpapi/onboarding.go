package httpapi

import (
	"database/sql"
	"errors"
	"strings"

	"ganigeorgiev.com/pocketbase/dbx"
	"github.com/bridge-allah/a-r-s-y/core"
)

type OnboardingStatus struct {
	SuperuserExists     bool `json:"superuser_exists"`
	SecretKeyConfigured bool `json:"secret_key_configured"`
	ServerRegistered    bool `json:"server_registered"`
}

func readOnboardingStatus(app core.App, secretKeyEnv string, secretLookup func(string) (string, bool)) (OnboardingStatus, error) {
	var status OnboardingStatus
	superuserCount, err := app.CountRecords(core.CollectionNameSuperusers, dbx.Not(dbx.HashExp{"email": core.DefaultInstallerEmail}))
	if err != nil {
		return status, err
	}
	status.SuperuserExists = superuserCount > 0

	secret, ok := secretLookup(secretKeyEnv)
	status.SecretKeyConfigured = ok && strings.TrimSpace(secret) != ""

	servers, err := app.FindCollectionByNameOrId("servers")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return status, nil
		}
		return status, err
	}
	serverCount, err := app.CountRecords(servers)
	if err != nil {
		return status, err
	}
	status.ServerRegistered = serverCount > 0

	return status, nil
}
