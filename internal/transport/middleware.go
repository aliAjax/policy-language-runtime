package transport

import (
	"context"
	"net/http"
	"time"
)

type Limit struct {
	Next http.Handler
	Max  int64
}

func (l Limit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, l.Max)
	l.Next.ServeHTTP(w, r)
}

type Timeout struct {
	Next     http.Handler
	Duration time.Duration
}

func (t Timeout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if t.Duration <= 0 {
		t.Next.ServeHTTP(w, r)
		return
	}
	ctx, c := context.WithTimeout(r.Context(), t.Duration)
	defer c()
	t.Next.ServeHTTP(w, r.WithContext(ctx))
}
