package manifest

import "testing"

func TestManifestValidate(t *testing.T) {
	tests := []struct {
		name    string
		m       Manifest
		wantErr bool
	}{
		{
			name: "valid minimal manifest",
			m: Manifest{
				SchemaVersion: CurrentSchemaVersion,
				Name:          "base",
				Version:       "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "valid full manifest",
			m: Manifest{
				SchemaVersion: CurrentSchemaVersion,
				Name:          "base.company",
				Version:       "0.1.0-alpha",
				Title:         "Base Company",
				Depends:       []Dependency{{Name: "base", Constraint: ">=1.0.0"}},
				Models:        []string{"models/company.yaml"},
				Data:          []string{"data/seed.yaml"},
				Security:      []string{"security/rules.yaml"},
				Views:         []string{"views/form.yaml"},
				Migrations:    []string{"migrations/001_init.yaml"},
			},
			wantErr: false,
		},
		{
			name: "missing schema version",
			m: Manifest{
				Name:    "base",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "wrong schema version",
			m: Manifest{
				SchemaVersion: "2.0",
				Name:          "base",
				Version:       "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "invalid module name",
			m: Manifest{
				SchemaVersion: CurrentSchemaVersion,
				Name:          "Base-Company",
				Version:       "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "invalid version",
			m: Manifest{
				SchemaVersion: CurrentSchemaVersion,
				Name:          "base",
				Version:       "v1",
			},
			wantErr: true,
		},
		{
			name: "path escapes module directory",
			m: Manifest{
				SchemaVersion: CurrentSchemaVersion,
				Name:          "base",
				Version:       "1.0.0",
				Models:        []string{"../outside.yaml"},
			},
			wantErr: true,
		},
		{
			name: "absolute path rejected windows",
			m: Manifest{
				SchemaVersion: CurrentSchemaVersion,
				Name:          "base",
				Version:       "1.0.0",
				Models:        []string{"C:\\\\Windows\\\\System32\\\\file.yaml"},
			},
			wantErr: true,
		},
		{
			name: "invalid dependency name",
			m: Manifest{
				SchemaVersion: CurrentSchemaVersion,
				Name:          "base",
				Version:       "1.0.0",
				Depends:       []Dependency{{Name: "INVALID"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
