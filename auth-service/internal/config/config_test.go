package config

import (
	"os"
	"testing"
)

// unset helper to redo all envs
func unset(keys ...string) {
	for _, k := range keys {
		_ = os.Unsetenv(k)
	}
}

// TestLoadConfig_Defaults_AreDevEnv test with default config
func TestLoadConfig_Defaults_AreDevEnv(t *testing.T) {
	unset("AUTH_ENV")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if cfg.Env != EnvProd {
		t.Fatalf("env=%q, want %q", cfg.Env, EnvProd)
	}

	if cfg.HTTP.Addr != DefaultHttpAddr {
		t.Fatalf("addr=%q, want %s", cfg.HTTP.Addr, DefaultHttpAddr)
	}

	if cfg.Observability.MetricsConfig.Enabled {
		t.Fatalf("metrics should default to false")
	}
}

// TestLoadConfig_EnvOverride checks overrides env with default
func TestLoadConfig_EnvOverride(t *testing.T) {
	_ = os.Setenv("AUTH_HTTP__ADDR", ":9090")
	defer unset("AUTH_HTTP__ADDR")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("addr=%q, want :9090", cfg.HTTP.Addr)
	}
}
