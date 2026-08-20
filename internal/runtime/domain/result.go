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
	if values, ok := r.Value.([]string); ok {
		r.Value = append([]string(nil), values...)
	}
	return r
}
