package domain

import "errors"

var ErrRegistryUnavailable = errors.New("")

type Signature struct {
	Name          string
	Args          []string
	Returns       string
	Deterministic bool
}
type Registry interface {
	Register(Signature) error
	Lookup(string) (Signature, bool)
	Ready() bool
}
