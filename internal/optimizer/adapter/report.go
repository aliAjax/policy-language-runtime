package adapter

type Report struct {
	ConstantsFolded int
	BranchesRemoved int
	Dependencies    int
	DependencyNames []string
}

func (r Report) Healthy() bool { return r.ConstantsFolded >= 0 && r.BranchesRemoved >= 0 }
func (r Report) Clone() Report {
	return Report{
		ConstantsFolded: r.ConstantsFolded,
		BranchesRemoved: r.BranchesRemoved,
		Dependencies:    r.Dependencies,
		DependencyNames: append([]string(nil), r.DependencyNames...),
	}
}
