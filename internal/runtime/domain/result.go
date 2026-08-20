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
	return r
}
