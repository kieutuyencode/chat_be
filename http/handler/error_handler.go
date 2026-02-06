package handler

import (
	"backend/apperror"
	"backend/common/result"
	"backend/logger"
	"net/http"
	"runtime/debug"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ErrorHandler iris.Handler

type errorHandlerParams struct {
	fx.In
	fx.Lifecycle
	Logger *logger.Logger
}

func newErrorHandler(p errorHandlerParams) ErrorHandler {
	return func(ctx iris.Context) {
		defer func() {
			if r := recover(); r != nil {
				p.Logger.Error(string(logger.MessageRequestFailed), zap.Any("panic", r), zap.ByteString("stacktrace", debug.Stack()))

				ctx.StatusCode(http.StatusInternalServerError)
				ctx.JSON(result.Fail(http.StatusText(http.StatusInternalServerError), nil))
			}

			if err := ctx.GetErr(); err != nil {
				var appErr *apperror.AppError
				var validationErrors validator.ValidationErrors

				if errors.As(err, &appErr) {
					p.Logger.Warn(string(logger.MessageRequestFailed), zap.String("message", appErr.Message), zap.Error(err))

					res := result.Fail(appErr.Message, appErr.Data)

					switch appErr.Code {
					case apperror.CodeNotFound:
						ctx.StatusCode(http.StatusNotFound)
					case apperror.CodeBadRequest:
						ctx.StatusCode(http.StatusBadRequest)
					case apperror.CodeUnauthorized:
						ctx.StatusCode(http.StatusUnauthorized)
					case apperror.CodeForbidden:
						ctx.StatusCode(http.StatusForbidden)
					}

					if errors.As(err, &validationErrors) {
						errorMessages := make(map[string]string)
						for _, fieldError := range validationErrors {
							errorMessages[fieldError.Field()] = "Failed on validation tag: '" + fieldError.Tag() + "=" + fieldError.Param() + "'"
						}

						res.Detail = errorMessages
					}

					ctx.JSON(res)
				} else {
					p.Logger.Error(string(logger.MessageRequestFailed), zap.Error(err))

					ctx.StatusCode(http.StatusInternalServerError)
					ctx.JSON(result.Fail(http.StatusText(http.StatusInternalServerError), nil))
				}
			}
		}()

		ctx.Next()
	}
}
