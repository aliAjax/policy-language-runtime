package adapter

import (
	"fmt"
	"github.com/example/policy-language-runtime/internal/typecheck/domain"
)

func Format(ds []domain.Diagnostic) string {
	out := ""
	for _, d := range ds {
		if !d.Valid() {
			continue
		}
		prefix := d.Severity
		if d.Code != "" {
			prefix += "[" + d.Code + "]"
		}
		out += fmt.Sprintf("%s: %s\n", prefix, d.Message)
	}
	return out
}
