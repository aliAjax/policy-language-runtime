package domain

type Decision string

const (
	Allow         Decision = "allow"
	Deny          Decision = "deny"
	Indeterminate Decision = "indeterminate"
)

type Result struct {
	Decision Decision
	Value    any
	Reason   string
	Version  string
	Steps    []string
}

func (r Result) Clone() Result {
	r.Steps = append([]string(nil), r.Steps...)
	switch value := r.Value.(type) {
	case []string:
		r.Value = append([]string(nil), value...)
	case []any:
		r.Value = append([]any(nil), value...)
	}
	return r
}
