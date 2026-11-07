package migrations

import (
	"database/sql"
	"errors"

	"github.com/bridge-allah/a-r-s-y/core"
)

const serversCollectionName = "servers"

func init() {
	core.SystemMigrations.Register(createServersCollection, deleteServersCollection)
}

func createServersCollection(txApp core.App) error {
	_, err := txApp.FindCollectionByNameOrId(serversCollectionName)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	collection := core.NewBaseCollection(serversCollectionName)
	collection.Fields.Add(&core.TextField{Name: "name", Required: true, Max: 120})
	collection.Fields.Add(&core.TextField{Name: "host", Required: true, Max: 253})
	collection.Fields.Add(&core.NumberField{Name: "ssh_port", Required: true, OnlyInt: true, Min: floatPointer(1), Max: floatPointer(65535)})
	collection.Fields.Add(&core.TextField{Name: "ssh_username", Required: true, Max: 64, Pattern: "^[a-z_][a-z0-9_-]*[$]?$"})
	collection.Fields.Add(&core.TextField{Name: "os_name", Max: 80})
	collection.Fields.Add(&core.TextField{Name: "os_version", Max: 120})
	collection.Fields.Add(&core.TextField{Name: "host_key_fingerprint", Max: 128})
	collection.Fields.Add(&core.SelectField{Name: "readiness_state", Values: []string{"unknown", "pending", "ready", "unreachable", "failed"}})
	collection.Fields.Add(&core.TextField{Name: "private_key_ciphertext", Hidden: true, Max: 20000})
	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})

	return txApp.Save(collection)
}

func deleteServersCollection(txApp core.App) error {
	collection, err := txApp.FindCollectionByNameOrId(serversCollectionName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return txApp.Delete(collection)
}

func floatPointer(value float64) *float64 {
	return &value
}
