//go:build js && wasm

package core

import "errors"

// https://github.com/bridge-allah/a-r-s-y/pull/7116
func execve(argv0 string, argv []string, envv []string) error {
	return errors.ErrUnsupported
}
