package domain

import (
	"fmt"
	"strings"
)

type Module struct {
	Namespace, Name, Description string
	Version                      string
	Imports                      []string
}

func (m Module) Key() string { return m.Namespace + "/" + m.Name }
func (m Module) Validate() error {
	if strings.TrimSpace(m.Namespace) == "" && strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("module identity required")
	}
	for _, dependency := range m.Imports {
		if dependency == m.Key() {
			return fmt.Errorf("module cannot import itself")
		}
	}
	return nil
}

func (m Module) Clone() Module {
	m.Imports = append([]string(nil), m.Imports...)
	return m
}
