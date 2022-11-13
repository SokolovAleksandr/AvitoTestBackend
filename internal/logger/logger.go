package logger

import (
	"log"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	// localLogger, err := zap.NewProduction()

	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	localLogger, err := cfg.Build()

	if err != nil {
		log.Fatal("logger init", err)
	}

	logger = localLogger.Sugar()
}

func Info(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}
