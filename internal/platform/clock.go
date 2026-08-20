package platform

import "time"

type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}
type SystemClock struct{}

func (SystemClock) Now() time.Time                  { return time.Now() }
func (SystemClock) Since(t time.Time) time.Duration { return time.Since(t) }
