package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr        string
	MaxBody     int64
	ExecTimeout time.Duration
	MaxSteps    int
}

func Default() Config {
	return Config{Addr: ":8095", MaxBody: 1 << 20, ExecTimeout: 5 * time.Second, MaxSteps: 10000}
}
func Load() Config {
	c := Default()
	path := os.Getenv("POLICY_CONFIG")
	if path == "" {
		path = "configs/config.yaml"
	}
	_ = loadYAML(path, &c)
	if v := os.Getenv("POLICY_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("POLICY_MAX_BODY"); v != "" {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil {
			c.MaxBody = n
		}
	}
	if v := os.Getenv("POLICY_MAX_STEPS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			c.MaxSteps = n
		}
	}
	if v := os.Getenv("POLICY_EXEC_TIMEOUT"); v != "" {
		if d, e := time.ParseDuration(v); e == nil {
			c.ExecTimeout = d
		}
	}
	return c
}

// loadYAML handles the flat scalar configuration used by the service and keeps
// the standard library-only runtime small; environment variables remain the
// highest-precedence override.
func loadYAML(path string, c *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.SplitN(strings.TrimSpace(s.Text()), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		switch key {
		case "addr":
			c.Addr = value
		case "max_body":
			if n, e := strconv.ParseInt(value, 10, 64); e == nil {
				c.MaxBody = n
			}
		case "max_steps":
			if n, e := strconv.Atoi(value); e == nil {
				c.MaxSteps = n
			}
		case "exec_timeout":
			if d, e := time.ParseDuration(value); e == nil {
				c.ExecTimeout = d
			}
		}
	}
	return s.Err()
}
func (c Config) Validate() error {
	if c.Addr == "" || c.MaxBody <= 0 || c.MaxSteps <= 0 || c.ExecTimeout <= 0 {
		return fmt.Errorf("invalid config")
	}
	return nil
}
