package domain

type Update struct {
	Namespace, Module, Version, Channel, Checksum string
	Signature                                     string
	Payload                                       []byte
	Cursor                                        int64
}
type Cursor struct{ Value int64 }

func (u Update) Clone() Update {
	c := u
	if u.Payload != nil {
		payload := make([]byte, len(u.Payload))
		copy(payload, u.Payload)
		c.Payload = payload
	}
	return c
}
