package apperror

import (
	"backend/logger"
	"runtime/debug"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Handler func(fn func() error) error

type handlerParams struct {
	fx.In
	Logger *logger.Logger
}

func newHandler(p handlerParams) Handler {
	return func(fn func() error) error {
		defer func() {
			if r := recover(); r != nil {
				p.Logger.Error(string(logger.MessageHandlerFailed), zap.Any("panic", r), zap.ByteString("stacktrace", debug.Stack()))
			}
		}()

		err := fn()
		if err != nil {
			p.Logger.Error(string(logger.MessageHandlerFailed), zap.Error(err))
		}

		return err
	}
}
