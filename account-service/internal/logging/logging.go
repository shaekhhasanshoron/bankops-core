package logging

import (
	"account-service/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"io"
	"os"
)

var Logger zerolog.Logger

func InitiateLogger() error {
	var logLevel zerolog.Level
	level := config.Current().Logging.Level

	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var writer io.Writer
	if config.Current().Logging.Encoding == "json" {
		writer = os.Stdout
	} else {
		writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
	}

	Logger = zerolog.New(writer).
		Level(logLevel).
		With().
		Timestamp().
		Str("service", config.ServiceName).
		Logger()

	// Include stack traces for errors in development
	if config.Current().Env != config.EnvProd {
		Logger = Logger.With().Stack().Caller().Logger()
	}
	return nil
}
