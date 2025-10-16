/*
	Centralized configuration.

	Load order & precedence (lowest → highest):
		- defaults -> .env files (loaded when AUTH_ENV != "prod") -> *_FILE file injection -> real environment variables

	Env naming conventions:
		- Use a service-specific prefix: AUTH_.
		- Use double underscores "__" to denote nesting. e.g. AUTH_HTTP__ADDR=":8080"  -> http.addr
	Files Paths (*_FILE):
		- This env denotes the filePath of any secret/private data; the files can be injected by any external source e.g. Vault.
*/

package config

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	ServiceName = "auth-service"
	EnvDev      = "dev"
	EnvStaging  = "staging"
	EnvProd     = "prod"

	DefaultHttpAddr = ":8080"
	DefaultGRPCAddr = ":50051"

	// EnvPrefix keeps unique environment prefix
	EnvPrefix = "AUTH_"

	// FileSuffix for "value from file" envs (Vault/K8s/Docker secrets pattern).
	FileSuffix = "_FILE"
)

type Config struct {
	Env           string           `koanf:"env" validate:"required,oneof=dev staging prod"`
	Auth          AuthConfig       `koanf:"auth" validate:"required"`
	User          UserConfig       `koanf:"user" validate:"required"`
	GRPC          GrpcConfig       `koanf:"grpc" validate:"required"`
	HTTP          HTTPConfig       `koanf:"http" validate:"required"`
	Logging       LoggingCfg       `koanf:"logging" validate:"required"`
	Observability ObservabilityCfg `koanf:"observability" validate:"required"`
	DB            DBConfig         `koanf:"db" validate:"required"`
}

type AuthConfig struct {
	HashKey           string        `koanf:"hash_key"`
	JWTSecret         string        `koanf:"jwt_secret"`
	JWTTokentDuration time.Duration `koanf:"jwt_token_duration"`
}

type UserConfig struct {
	AdminUsername string `koanf:"admin_username"`
	AdminPassword string `koanf:"admin_password"`
}

type GrpcConfig struct {
	Addr string `koanf:"addr"                  validate:"required"`
}

type HTTPConfig struct {
	Addr                string `koanf:"addr"                  validate:"required"`
	ReadTimeoutSeconds  int    `koanf:"read_timeout_seconds"  validate:"gte=1,lte=120"`
	WriteTimeoutSeconds int    `koanf:"write_timeout_seconds" validate:"gte=1,lte=120"`
	IdleTimeoutSeconds  int    `koanf:"idle_timeout_seconds"  validate:"gte=1,lte=300"`
}

type LoggingCfg struct {
	Level    string `koanf:"level"    validate:"required,oneof=debug info warn error"`
	Encoding string `koanf:"encoding" validate:"required,oneof=json console"`
}

type ObservabilityCfg struct {
	MetricsConfig MetricsCfg `koanf:"metrics"`
	TracingConfig TracingCfg `koanf:"tracing"`
}

type MetricsCfg struct {
	Enabled bool `koanf:"enabled"`
}

type TracingCfg struct {
	Enabled  bool   `koanf:"enabled"`
	Protocol string `koanf:"protocol"    validate:"oneof=grpc http"`
	Endpoint string `koanf:"endpoint"`
}

type DBConfig struct {
	DSN  string `koanf:"dsn"`
	Type string `koanf:"type"`
}

var (
	global     Config
	globalOnce sync.Once
)

func setGlobal(c Config) {
	globalOnce.Do(func() {
		global = c
	})
}

func Current() Config { return global }

// LoadConfig builds configuration from defaults, optional .env, *_FILE secrets, and process env.
func LoadConfig(envFiles ...string) (*Config, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// Loads initial variables
	k := koanf.New(".")

	if err := k.Load(confmap.Provider(defaults(), "."), nil); err != nil {
		logger.Fatal().Err(err).Msg("could not load default env variables")
	}

	// Setting default env
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_ENV")))
	if appEnv == "" {
		appEnv = EnvDev
	}

	if appEnv != EnvProd {
		if len(envFiles) > 0 {
			if err := godotenv.Load(envFiles...); err != nil {
				logger.Fatal().Err(err).Msg(".env load")
			}
		} else {
			_ = godotenv.Load()
		}

		// validating from .env file again and overwriting if exists
		appEnvOverwrite := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_ENV")))
		if appEnvOverwrite != "" {
			appEnv = appEnvOverwrite
		}
	}

	// Resolve env from files (env name ends with `_FILE` or `FILE`)
	if err := injectFiles(k, EnvPrefix, FileSuffix); err != nil {
		logger.Fatal().Err(err).Msg("file validation failed")
	}

	if err := k.Load(env.Provider(EnvPrefix, ".", func(s string) string {
		key := strings.TrimPrefix(s, EnvPrefix)
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, "__", ".")
		return key
	}), nil); err != nil {
		logger.Fatal().Err(err).Msg("could not load env variables")
	}

	// Convert env to struct
	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal main config")
	}

	// validating the config data
	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal main config")
	}

	// Setting environment globally
	setGlobal(*cfg)

	logger.Debug().
		Str("env", cfg.Env).
		Msg("Environment Variables Initialized")
	return cfg, nil
}

// injectFiles reads AUTH_*_FILE envs and sets their contents to the corresponding key.
// e.g. AUTH_DB__PASSWORD_FILE=/config/secret/data → (get data from file /config/secret/data) -> sets "db.password"  (if AUTH_DB__PASSWORD is not set in env directly).
func injectFiles(k *koanf.Koanf, prefix, suffix string) error {
	var errs []string

	// read all environments
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envKey, filePath := parts[0], parts[1] // e.g. envKey=AUTH_DB__PASSWORD_FILE; filePath=/config/secret/data

		if !strings.HasPrefix(envKey, prefix) || !strings.HasSuffix(envKey, suffix) {
			continue
		}

		base := strings.TrimSuffix(envKey, suffix) // e.g. AUTH_DB__PASSWORD_FILE -> AUTH_DB__PASSWORD

		// If key already exists in environment then skip
		if _, ok := os.LookupEnv(base); ok {
			continue
		}

		// Checking file existence
		b, err := os.ReadFile(filepath.Clean(strings.TrimSpace(filePath)))
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, fmt.Sprintf("%s: %v", envKey, err))
			}
			continue
		}

		// maps env -> koanf key; e.g. AUTH_DB__PASSWORD -> db.password
		mapped := strings.ToLower(strings.TrimPrefix(base, prefix))
		mapped = strings.ReplaceAll(mapped, "__", ".")

		if len(b) > 1<<20 { // 1 MiB safety
			errs = append(errs, fmt.Sprintf("%s: secret file too large", envKey))
			continue
		}

		val := strings.TrimRight(string(b), "\r\n\t ")
		if val == "" {
			errs = append(errs, fmt.Sprintf("%s: secret file is empty", envKey))
			continue
		}

		_ = k.Set(mapped, val)
	}

	if len(errs) > 0 {
		return errors.New("injected file load errors: " + strings.Join(errs, "; "))
	}
	return nil
}

func defaults() map[string]any {
	return map[string]any{
		"env": EnvProd,
		"grpc": map[string]any{
			"addr": DefaultGRPCAddr,
		},
		"http": map[string]any{
			"addr":                  DefaultHttpAddr,
			"read_timeout_seconds":  15,
			"write_timeout_seconds": 15,
			"idle_timeout_seconds":  120,
		},
		"logging": map[string]any{
			"level":    "info",
			"encoding": "console",
		},
		"observability": map[string]any{
			"metrics": map[string]any{
				"enabled": false,
			},
			"tracing": map[string]any{
				"enabled":  false,
				"protocol": "grpc",
				"endpoint": "localhost:4317",
			},
		},
		"db": map[string]any{
			"dsn":  "./auth-service.db",
			"type": "sqlite",
		},
		"user": map[string]any{
			"admin_username": "admin",
			"admin_password": "admin",
		},
		"auth": map[string]any{
			"hash_key":           "fc5c6816998c7173ba5bc7a3c53bfabf",
			"jwt_secret":         "fc5c6816998c7173ba5bc7a3c53bfabf",
			"jwt_token_duration": 10 * time.Minute,
		},
	}
}
