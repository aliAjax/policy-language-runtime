package domain

import "fmt"

type State string

const (
	Draft      State = "draft"
	Active     State = "active"
	Suspended  State = "suspended"
	Recovering State = "recovering"
	Archived   State = "archived"
)

type Namespace struct {
	ID, Name, Owner string
	Active          bool
	State           State
}

func (n Namespace) Valid() bool {
	return n.ID != "" && n.Name != "" && (n.State == "" || validState(n.State))
}

func (n Namespace) CurrentState() State {
	if n.State != "" {
		return n.State
	}
	if n.Active {
		return Active
	}
	return Draft
}

func (n *Namespace) Transition(next State) error {
	current := n.CurrentState()
	if !CanTransition(current, next) {
		return fmt.Errorf("namespace transition %s -> %s is not allowed", current, next)
	}
	n.State = next
	n.Active = next == Active
	return nil
}

func CanTransition(from, to State) bool {
	allowed := map[State]map[State]bool{
		Draft:      {Active: true, Archived: true},
		Active:     {Draft: true, Suspended: true, Archived: true},
		Suspended:  {Recovering: true, Archived: true},
		Recovering: {Suspended: true},
		Archived:   {},
	}
	return validState(from) && validState(to) && allowed[from][to]
}

func validState(state State) bool {
	switch state {
	case Draft, Active, Suspended, Recovering, Archived:
		return true
	default:
		return false
	}
}
