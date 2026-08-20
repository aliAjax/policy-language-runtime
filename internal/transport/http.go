package transport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/example/policy-language-runtime/internal/platform"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	svc    *platform.Service
	logger *slog.Logger
}

func New(s *platform.Service, l *slog.Logger) *Server { return &Server{svc: s, logger: l} }
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte(`{"status":"ok"}`)) })
	m.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte(`{"ready":true}`)) })
	m.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("policy_requests_total 0\n"))
	})
	m.HandleFunc("/v1/compile", s.compile)
	m.HandleFunc("/v1/publish", s.publish)
	m.HandleFunc("/v1/decide", s.decide)
	m.HandleFunc("/v1/explain", s.decide)
	return middleware(m, s.logger)
}
func (s *Server) compile(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Source string `json:"source"`
	}
	if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
		writeErr(w, e)
		return
	}
	p, d, e := s.svc.Compile(r.Context(), in.Source)
	if e != nil {
		writeJSON(w, 400, map[string]any{"error": e.Error(), "diagnostics": d})
		return
	}
	writeJSON(w, 200, map[string]any{"compiled": true, "program": p})
}
func (s *Server) publish(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Namespace string `json:"namespace"`
		Module    string `json:"module"`
		Version   string `json:"version"`
		Channel   string `json:"channel"`
		Source    string `json:"source"`
	}
	if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
		writeErr(w, e)
		return
	}
	if in.Namespace == "" {
		in.Namespace = "default"
	}
	if in.Module == "" {
		in.Module = "main"
	}
	if in.Version == "" {
		in.Version = "1.0.0"
	}
	if in.Channel == "" {
		in.Channel = "stable"
	}
	p, e := s.svc.Publish(r.Context(), in.Namespace, in.Module, in.Version, in.Channel, in.Source)
	if e != nil {
		writeErr(w, e)
		return
	}
	writeJSON(w, 200, p)
}
func (s *Server) decide(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Namespace string         `json:"namespace"`
		Module    string         `json:"module"`
		Version   string         `json:"version"`
		Channel   string         `json:"channel"`
		Input     map[string]any `json:"input"`
	}
	if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
		writeErr(w, e)
		return
	}
	if in.Namespace == "" {
		in.Namespace = "default"
	}
	if in.Module == "" {
		in.Module = "main"
	}
	res, e := s.svc.Decide(r.Context(), in.Namespace, in.Module, in.Version, in.Channel, in.Input)
	if e != nil {
		writeErr(w, e)
		return
	}
	writeJSON(w, 200, res)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, e error) {
	status := http.StatusBadRequest
	code := "invalid_request"
	switch {
	case errors.Is(e, platform.ErrNotFound):
		status = http.StatusInternalServerError
		code = "internal_error"
	case errors.Is(e, platform.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(e, context.DeadlineExceeded):
		status, code = http.StatusGatewayTimeout, "deadline_exceeded"
	case errors.Is(e, context.Canceled):
		status, code = 499, "request_canceled"
	}
	writeJSON(w, status, map[string]string{"code": code, "error": e.Error()})
}
func middleware(next http.Handler, l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = time.Now().UTC().Format("20060102150405.000000000")
		}
		w.Header().Set("X-Request-ID", id)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		l.Info("request", "request_id", id, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

var _ = strings.TrimSpace
