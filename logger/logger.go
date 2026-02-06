package logger

import (
	"context"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.Logger
}

type params struct {
	fx.In
	fx.Lifecycle
}

func new(p params) *Logger {
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename: "logs/app.log",
		MaxSize:  100, // MB
	})

	stdout := zapcore.AddSync(os.Stdout)

	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	fileEncoder := zapcore.NewJSONEncoder(productionCfg)
	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, file, level),
		zapcore.NewCore(consoleEncoder, stdout, level),
	)

	logger := zap.New(core)

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			defer logger.Sync()
			return nil
		},
	})

	return &Logger{Logger: logger}
}
