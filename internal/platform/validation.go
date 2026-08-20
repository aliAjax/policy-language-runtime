package platform

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	MaxSource       int
	MaxVersion      int
	AllowedChannels map[string]bool
}

func NewValidator() *Validator {
	return &Validator{MaxSource: 1 << 20, MaxVersion: 32, AllowedChannels: map[string]bool{"stable": true, "canary": true, "next": true}}
}
func (v *Validator) Source(src string) error {
	if src == "" {
		return fmt.Errorf("source is empty")
	}
	if len(src) > v.MaxSource {
		return fmt.Errorf("source exceeds %d bytes", v.MaxSource)
	}
	if !utf8.ValidString(src) {
		return fmt.Errorf("source is not valid UTF-8")
	}
	return nil
}
func (v *Validator) Version(ver string) error {
	if ver == "" || len(ver) > v.MaxVersion {
		return fmt.Errorf("invalid version")
	}
	parts := strings.Split(ver, ".")
	if len(parts) != 3 {
		return fmt.Errorf("version must be semantic")
	}
	for _, p := range parts {
		if p == "" {
			return fmt.Errorf("empty version component")
		}
	}
	return nil
}
func (v *Validator) Channel(ch string) error {
	if !v.AllowedChannels[ch] {
		return fmt.Errorf("unsupported channel %s", ch)
	}
	return nil
}
func (v *Validator) Identity(ns, mod string) error {
	if strings.TrimSpace(ns) == "" || strings.TrimSpace(mod) == "" {
		return fmt.Errorf("namespace and module required")
	}
	if strings.ContainsAny(ns, "/\\") || strings.ContainsAny(mod, "/\\") {
		return fmt.Errorf("identity contains path separator")
	}
	return nil
}
func (v *Validator) Input(in map[string]any) error {
	if in == nil {
		return fmt.Errorf("input is nil")
	}
	if len(in) > 1024 {
		return fmt.Errorf("input field count exceeded")
	}
	return nil
}
func (v *Validator) All(ns, mod, ver, ch, src string) error {
	if e := v.Identity(ns, mod); e != nil {
		return e
	}
	if e := v.Version(ver); e != nil {
		return e
	}
	if e := v.Channel(ch); e != nil {
		return e
	}
	return v.Source(src)
}

type AuditEvent struct{ Action, Namespace, Module, Version, Actor, Reason string }

func (e AuditEvent) Valid() bool { return e.Action != "" && e.Namespace != "" && e.Module != "" }

type AuditLog struct{ Events []AuditEvent }

func (l *AuditLog) Append(e AuditEvent) error {
	if !e.Valid() {
		return fmt.Errorf("invalid audit event")
	}
	l.Events = append(l.Events, e)
	return nil
}
func (l AuditLog) ForModule(ns, mod string) []AuditEvent {
	o := []AuditEvent{}
	for _, e := range l.Events {
		if e.Namespace == ns && e.Module == mod {
			o = append(o, e)
		}
	}
	return o
}

type CursorStore struct{ Value int64 }

func (c *CursorStore) Advance(n int64) bool {
	if n <= c.Value {
		return false
	}
	c.Value = n
	return true
}
func (c CursorStore) Valid(n int64) bool { return n >= c.Value }

type Release struct {
	Version, Channel string
	Active, Paused   bool
	State            string
}

func (r Release) CanEvaluate() bool { return r.Active && !r.Paused && r.State == "active" }
func (r *Release) Pause()           { r.Paused = true }
func (r *Release) Resume()          { r.Paused = false }
func (r *Release) Activate()        { r.Active, r.State = true, "active" }
func (r *Release) Deactivate()      { r.Active, r.State = false, "suspended" }
func (r *Release) Recover() {
	r.State = "recovering"
	r.Paused = false
}

type DependencyGraph struct{ Edges map[string][]string }

func NewGraph() *DependencyGraph               { return &DependencyGraph{Edges: map[string][]string{}} }
func (g *DependencyGraph) Add(from, to string) { g.Edges[from] = append(g.Edges[from], to) }
func (g *DependencyGraph) Reachable(start string) []string {
	seen := map[string]bool{}
	out := []string{}
	var visit func(string)
	visit = func(n string) {
		if seen[n] {
			return
		}
		seen[n] = true
		out = append(out, n)
		for _, x := range g.Edges[n] {
			visit(x)
		}
	}
	visit(start)
	return out
}
func (g *DependencyGraph) HasCycle(start string) bool {
	active := map[string]bool{}
	done := map[string]bool{}
	var walk func(string) bool
	walk = func(n string) bool {
		if active[n] {
			return true
		}
		if done[n] {
			return false
		}
		active[n] = true
		for _, x := range g.Edges[n] {
			if walk(x) {
				return true
			}
		}
		delete(active, n)
		done[n] = true
		return false
	}
	return walk(start)
}
