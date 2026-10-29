//go:build !windows && !linux

package service

import "errors"

const Name = "nevarix-agent"

func Start() error   { return errUnsupported() }
func Stop() error    { return errUnsupported() }
func Restart() error { return errUnsupported() }
func Delete() error  { return errUnsupported() }
func Create(_, _ string) error { return errUnsupported() }
func PersistEnv(map[string]string) error { return errUnsupported() }

func Status() (string, error) { return "", errUnsupported() }

func errUnsupported() error {
	return errors.New("service install is supported on Windows and Linux only")
}
