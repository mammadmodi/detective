package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger is a factory function which creates a zap logger based on the entry config file.
func NewZapLogger(name string, config *Config) (*zap.Logger, error) {
	// Return a nop logger if logger is not enabled.
	if !config.Enabled {
		return zap.NewNop(), nil
	}

	// if Pretty flag in config is enabled use a ConsoleEncoder which is human readable.
	var encoder zapcore.Encoder
	if config.Pretty {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		encoder = zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	}

	zapLvl := parseLevel(config.Level)
	defaultCore := zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(os.Stdout)), zapLvl)
	cores := []zapcore.Core{defaultCore}

	// If FileRedirectEnabled flag in config is enabled add a core with a file writer to zap logger cores.
	if config.FileRedirectEnabled {
		// Check existence of file.
		fileName := fmt.Sprintf("%s/%s.log", config.FileRedirectPath, config.FileRedirectPrefix)
		basePath := filepath.Dir(fileName)
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("base path `%s` is not exist, error: %v", basePath, err)
		}
		// Open the file.
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return nil, fmt.Errorf("error while openning file `%s` for logging, err: %v", fileName, err)
		}

		// Append new core to zap logger cores.
		fileLoggerCore := zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(file)), zapLvl)
		cores = append(cores, fileLoggerCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)).Named(name)

	return logger, nil
}

// parseLevel will convert a string based log level to a zapcore.Level object.
func parseLevel(lvl string) zapcore.Level {
	lvl = strings.ToLower(lvl)
	switch lvl {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.ErrorLevel
	}
}
