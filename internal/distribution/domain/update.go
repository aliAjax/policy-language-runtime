package domain

type Update struct {
	Namespace, Module, Version, Channel, Checksum string
	Signature                                     string
	Payload                                       []byte
	Cursor                                        int64
}
type Cursor struct{ Value int64 }

func (u Update) Clone() Update {
	u.Payload = append([]byte(nil), u.Payload...)
	return u
}
