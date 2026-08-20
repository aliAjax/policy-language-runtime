package platform

type Health struct {
	Service, Version string
	Ready            bool
	Checks           map[string]string
}

func NewHealth(v string) *Health {
	return &Health{Service: "policyd", Version: v, Ready: true, Checks: map[string]string{}}
}
func (h *Health) Set(name, status string) {
	if h.Checks == nil {
		h.Checks = map[string]string{}
	}
	h.Checks[name] = status
}
func (h Health) Healthy() bool {
	if !h.Ready {
		return false
	}
	for _, v := range h.Checks {
		if v != "ok" {
			return false
		}
	}
	return true
}
