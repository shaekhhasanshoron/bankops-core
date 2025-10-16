package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
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
	unset("ACCOUNT_ENV")
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
	_ = os.Setenv("ACCOUNT_HTTP__ADDR", ":9090")
	defer unset("ACCOUNT_HTTP__ADDR")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("addr=%q, want :9090", cfg.HTTP.Addr)
	}
}

// TestLoadConfig_SuccessWithCustomEnvFile tests with temporary config file
func TestLoadConfig_SuccessWithCustomEnvFile(t *testing.T) {
	envContent := `
ACCOUNT_ENV=staging
ACCOUNT_GRPC__ADDR=:9090
ACCOUNT_HTTP__ADDR=:9091
`
	tmpFile, err := os.CreateTemp("", "test_env_*.env")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(envContent)
	assert.NoError(t, err)
	tmpFile.Close()

	config, err := LoadConfig(tmpFile.Name())

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "staging", config.Env)
	assert.Equal(t, ":9090", config.GRPC.Addr)
	assert.Equal(t, ":9091", config.HTTP.Addr)
}

// TestInjectFiles_ErrorFileNotFound tests if provided file not found
func TestInjectFiles_ErrorFileNotFound(t *testing.T) {
	k := koanf.New(".")

	originalEnv := os.Getenv("ACCOUNT_MISSING_FILE")
	defer os.Setenv("ACCOUNT_MISSING_FILE", originalEnv)

	os.Setenv("ACCOUNT_MISSING_FILE", "/non/existent/file")

	err := injectFiles(k, "ACCOUNT_", "_FILE")
	assert.NoError(t, err)
}
