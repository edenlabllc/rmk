package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	encoder      zapcore.Encoder
	stdoutSyncer zapcore.WriteSyncer
	stderrSyncer zapcore.WriteSyncer
	cores        []zapcore.Core
	Format       string
	Level        string
}

func debugLevel() zap.LevelEnablerFunc {
	return func(level zapcore.Level) bool {
		return level == zapcore.DebugLevel
	}
}

func infoLevel() zap.LevelEnablerFunc {
	return func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel || level == zapcore.WarnLevel
	}
}

func errorLevel() zap.LevelEnablerFunc {
	return func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel || level == zapcore.FatalLevel
	}
}

func (l *Logger) debugCores() []zapcore.Core {
	return append(l.cores,
		zapcore.NewCore(l.encoder, l.stdoutSyncer, debugLevel()),
		zapcore.NewCore(l.encoder, l.stdoutSyncer, infoLevel()),
		zapcore.NewCore(l.encoder, l.stderrSyncer, errorLevel()),
	)
}

func (l *Logger) infoCores() []zapcore.Core {
	return append(l.cores,
		zapcore.NewCore(l.encoder, l.stdoutSyncer, infoLevel()),
		zapcore.NewCore(l.encoder, l.stderrSyncer, errorLevel()),
	)
}

func (l *Logger) errorCores() []zapcore.Core {
	return append(l.cores, zapcore.NewCore(l.encoder, l.stderrSyncer, errorLevel()))
}

func Init(format, level string) func() {
	var logError error

	l := &Logger{
		stderrSyncer: zapcore.Lock(os.Stderr),
		stdoutSyncer: zapcore.Lock(os.Stdout),
		Format:       format,
		Level:        level,
	}

	cfgConsole := zapcore.EncoderConfig{
		MessageKey: "message",
		LevelKey:   "level",
		TimeKey:    "time",
		//CallerKey:     "caller",
		EncodeLevel: zapcore.CapitalColorLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
		//EncodeCaller:  zapcore.ShortCallerEncoder,
	}

	switch l.Format {
	case "json":
		cfgJson := cfgConsole
		cfgJson.EncodeLevel = zapcore.LowercaseLevelEncoder
		l.encoder = zapcore.NewJSONEncoder(cfgJson)
	case "console":
		l.encoder = zapcore.NewConsoleEncoder(cfgConsole)
	default:
		l.encoder = zapcore.NewConsoleEncoder(cfgConsole)
		logError = fmt.Errorf("invalid log output format %s, available: console, json", l.Format)
	}

	switch l.Level {
	case "debug":
		l.cores = l.debugCores()
	case "info", "warning":
		l.cores = l.infoCores()
	case "error":
		l.cores = l.errorCores()
	default:
		l.cores = l.errorCores()
		logError = fmt.Errorf("invalid log level %s, available: debug, info, error", l.Level)
	}

	// tee core
	core := zapcore.NewTee(l.cores...)

	// finally construct the logger with the tee core
	logger := zap.New(core, zap.AddCaller())
	undo := zap.ReplaceGlobals(logger)

	if logError != nil {
		zap.S().Fatal(logError)
	}

	return undo
}
