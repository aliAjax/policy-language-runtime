package domain

type Type string

const (
	Any    Type = "any"
	Bool   Type = "bool"
	Number Type = "number"
	String Type = "string"
	Null   Type = "null"
)

type Diagnostic struct {
	Message, Path string
	Code          string
	Line, Column  int
	Severity      string
}

func (d Diagnostic) Valid() bool {
	return d.Message != "" && d.Severity != ""
}
