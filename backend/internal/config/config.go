// Package config centralizes runtime settings so nothing important is
// hardcoded. Everything has a sane default for the local `make` flow, and can
// be overridden by environment for real deployments.
package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// UseEtcdRedis switches on the centralized backend (etcd peer registry +
	// Redis content index). When false (the default), the node runs purely on
	// libp2p — no external infra required.
	UseEtcdRedis  bool
	EtcdEndpoints []string
	RedisAddr     string
	BasePort      int
	MaxUploadSize int64    // bytes; rejects larger uploads
	CORSOrigins   []string // exact-match allowlist
}

const (
	defaultRedisAddr     = "localhost:6379"
	defaultEtcd          = "localhost:2379"
	defaultBasePort      = 5001
	defaultMaxUploadSize = 100 << 20 // 100 MiB
	defaultCORSOrigins   = "http://localhost:5173"
)

// Load reads configuration from the environment, falling back to defaults.
func Load() Config {
	return Config{
		UseEtcdRedis:  envBool("PHILE_USE_ETCD_REDIS", false),
		EtcdEndpoints: splitCSV(env("ETCD_ENDPOINTS", defaultEtcd)),
		RedisAddr:     env("REDIS_ADDR", defaultRedisAddr),
		BasePort:      envInt("BASE_PORT", defaultBasePort),
		MaxUploadSize: int64(envInt("MAX_UPLOAD_SIZE", defaultMaxUploadSize)),
		CORSOrigins:   splitCSV(env("CORS_ORIGINS", defaultCORSOrigins)),
	}
}

func envBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
