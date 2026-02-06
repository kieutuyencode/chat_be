package handler

import (
	"backend/http/key"
	"backend/logger"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RequestTracking iris.Handler

type requestTrackingParams struct {
	fx.In
	Logger *logger.Logger
}

func newRequestTracking(p requestTrackingParams) RequestTracking {
	return RequestTracking(func(ctx iris.Context) {
		requestId := uuid.New().String()
		requestCreatedAt := time.Now()
		requestPath := ctx.Request().URL.Path

		p.Logger.Info(
			string(logger.MessageRequestStarted),
			zap.String(key.Id, requestId),
			zap.String(key.Path, requestPath),
			zap.String(key.CreatedAt, requestCreatedAt.Format(time.RFC3339)),
			zap.String(key.Ip, ctx.RemoteAddr()),
		)

		ctx.Next()

		var clientId string
		if val := ctx.Values().Get(key.ClientId); val != nil {
			clientId = val.(string)
		}

		durationSeconds := time.Now().Sub(requestCreatedAt).Seconds()

		p.Logger.Info(
			string(logger.MessageRequestCompleted),
			zap.String(key.Id, requestId),
			zap.String(key.Path, requestPath),
			zap.Float64(key.DurationSeconds, durationSeconds),
			zap.String(key.ClientId, clientId),
		)
	})
}
