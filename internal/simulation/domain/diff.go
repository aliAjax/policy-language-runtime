package domain

type Sample struct {
	ID    string
	Input map[string]any
}

func (s Sample) Clone() Sample {
	s.Input = cloneInput(s.Input)
	return s
}

type Change struct {
	SampleID, Before, After string
	Kind                    string
}
type Report struct {
	Total   int
	Changes []Change
	Errors  int
	Samples []Sample
}

func (r Report) Clone() Report {
	r.Changes = append([]Change(nil), r.Changes...)
	samples := r.Samples
	r.Samples = make([]Sample, len(samples))
	for i := range samples {
		r.Samples[i] = samples[i].Clone()
	}
	return r
}

func cloneInput(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		switch typed := value.(type) {
		case []string:
			out[key] = append([]string(nil), typed...)
		case []any:
			out[key] = append([]any(nil), typed...)
		default:
			out[key] = value
		}
	}
	return out
}
