package adapter

import (
	"encoding/json"
	"github.com/example/policy-language-runtime/internal/ir/domain"
)

func Marshal(p domain.Program) ([]byte, error) { return json.Marshal(p) }
func Unmarshal(b []byte) (domain.Program, error) {
	var p domain.Program
	e := json.Unmarshal(b, &p)
	return p, e
}
