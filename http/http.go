package http

import (
	"backend/config"
	"backend/logger"
	"backend/websocket"
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/kataras/iris/v12"
	"github.com/philippseith/signalr"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type httpServerParams struct {
	fx.In
	fx.Lifecycle
	Router *iris.Application
	Logger *logger.Logger
	Env    *config.Env
	Server *websocket.Server
}

func newHTTPServer(p httpServerParams) (*http.Server, error) {
	if err := p.Router.Build(); err != nil {
		return nil, errors.Wrap(err, "failed to build Iris router")
	}

	mux := http.NewServeMux()
	mux.Handle("/", p.Router)
	p.Server.MapHTTP(signalr.WithHTTPServeMux(mux), "/websocket/v1")

	server := &http.Server{
		Addr:    ":" + p.Env.Port,
		Handler: mux,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", server.Addr)

			if err != nil {
				return errors.Wrapf(err, "failed to listen on %s", server.Addr)
			}

			p.Logger.Info(fmt.Sprintf("Starting HTTP server at http://localhost%s", server.Addr))

			go func() {
				if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
					p.Logger.Error(fmt.Sprintf("Failed to serve HTTP server %s", server.Addr), zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Shutdown HTTP server ...")

			if err := server.Shutdown(ctx); err != nil {
				p.Logger.Error("Failed to shutdown HTTP server", zap.Error(err))
			}

			p.Logger.Info("HTTP server exiting")

			return nil
		},
	})

	return server, nil
}
