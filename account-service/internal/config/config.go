/*
	Centralized configuration.

	Load order & precedence (lowest → highest):
		- defaults -> .env files (loaded when ACCOUNT_ENV != "prod") -> *_FILE file injection -> real environment variables

	Env naming conventions:
		- Use a service-specific prefix: ACCOUNT_.
		- Use double underscores "__" to denote nesting. e.g. ACCOUNT_HTTP__ADDR=":8080"  -> http.addr
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
	ServiceName     = "account-service"
	EnvDev          = "dev"
	EnvStaging      = "staging"
	EnvProd         = "prod"
	BrokerTypeKafka = "kafka"

	DefaultMessageBrokerMessagePublishTopic = "bank-events"
	DefaultMessageBrokerMessageEnabled      = false

	DefaultHttpAddr = ":8080"
	DefaultGRPCAddr = ":50051"

	// EnvPrefix keeps unique environment prefix
	EnvPrefix = "ACCOUNT_"

	// FileSuffix for "value from file" envs (Vault/K8s/Docker secrets pattern).
	FileSuffix = "_FILE"
)

type Config struct {
	Env              string                 `koanf:"env" validate:"required,oneof=dev staging prod"`
	Auth             AuthConfig             `koanf:"auth" validate:"required"`
	GRPC             GrpcConfig             `koanf:"grpc" validate:"required"`
	HTTP             HTTPConfig             `koanf:"http" validate:"required"`
	Logging          LoggingCfg             `koanf:"logging" validate:"required"`
	Observability    ObservabilityCfg       `koanf:"observability" validate:"required"`
	Recovery         RecoveryConfig         `koanf:"recovery" validate:"required"`
	Cleanup          CleanupConfig          `koanf:"cleanup" validate:"required"`
	DB               DBConfig               `koanf:"db" validate:"required"`
	MessagePublisher MessagePublisherConfig `koanf:"message_publisher" validate:"required"`
	AccountConfig    AccountConfig          `koanf:"account" validate:"required"`
}

type AccountConfig struct {
	MinDepositAmount float64 `koanf:"min_deposit_amount"  validate:"gte=0"`
}

type AuthConfig struct {
	HashKey string `koanf:"hash_key"`
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

type MessagePublisherConfig struct {
	Enabled      bool   `koanf:"enabled"`
	BrokerAddr   string `koanf:"broker_addr"`
	PublishTopic string `koanf:"publish_topic"`
	BrokerType   string `koanf:"broker_type"`
}

type RecoveryConfig struct {
	Enabled            bool          `koanf:"enabled"`
	Interval           time.Duration `koanf:"interval"`
	TransactionTimeout time.Duration `koanf:"transaction_timeout"`
	MaxRetries         int           `koanf:"max_retries"`
	StartupRecovery    bool          `koanf:"startup_recovery"`
}

type CleanupConfig struct {
	Enabled        bool          `koanf:"enabled"`
	Interval       time.Duration `koanf:"interval"`
	StaleThreshold time.Duration `koanf:"stale_threshold"`
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
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("ACCOUNT_ENV")))
	if appEnv == "" {
		appEnv = EnvProd
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
		appEnvOverwrite := strings.ToLower(strings.TrimSpace(os.Getenv("ACCOUNT_ENV")))
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
	return cfg, nil
}

// injectFiles reads ACCOUNT_*_FILE envs and sets their contents to the corresponding key.
// e.g. ACCOUNT_DB__PASSWORD_FILE=/config/secret/data → (get data from file /config/secret/data) -> sets "db.password"  (if ACCOUNT_DB__PASSWORD is not set in env directly).
func injectFiles(k *koanf.Koanf, prefix, suffix string) error {
	var errs []string

	// read all environments
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envKey, filePath := parts[0], parts[1] // e.g. envKey=ACCOUNT_DB__PASSWORD_FILE; filePath=/config/secret/data

		if !strings.HasPrefix(envKey, prefix) || !strings.HasSuffix(envKey, suffix) {
			continue
		}

		base := strings.TrimSuffix(envKey, suffix) // e.g. ACCOUNT_DB__PASSWORD_FILE -> ACCOUNT_DB__PASSWORD

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

		// maps env -> koanf key; e.g. ACCOUNT_DB__PASSWORD -> db.password
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
			"dsn":  "./account-service.db",
			"type": "sqlite",
		},
		"auth": map[string]any{
			"hash_key": "fc5c6816998c7173ba5bc7a3c53bfabf",
		},
		"recovery": map[string]any{
			"enabled":             true,
			"interval":            30 * time.Second,
			"transaction_timeout": 5 * time.Minute,
			"max_retries":         3,
			"startup_recovery":    true,
		},
		"cleanup": map[string]any{
			"enabled":         false,
			"interval":        1 * time.Hour,
			"stale_threshold": 24 * time.Hour,
		},
		"account": map[string]any{
			"min_deposit_amount": 0,
		},
		"message_publisher": map[string]any{
			"enabled":       DefaultMessageBrokerMessageEnabled,
			"broker_addr":   "",
			"publish_topic": "",
			"broker_type":   "",
		},
	}
}
