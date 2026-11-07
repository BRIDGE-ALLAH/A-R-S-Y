package manifest

import "errors"

var (
	ErrInvalidManifest   = errors.New("manifest: invalid manifest")
	ErrInvalidVersion    = errors.New("manifest: invalid version")
	ErrInvalidDependency = errors.New("manifest: invalid dependency")
	ErrInvalidPath       = errors.New("manifest: invalid file path")
	ErrUnknownKey        = errors.New("manifest: unknown key")
)
