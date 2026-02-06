package http

import (
	"backend/http/handler"
	"net/http"

	"go.uber.org/fx"
)

var Module = fx.Module("http",
	fx.Provide(newRouter, newHTTPServer),
	handler.Module,
	fx.Invoke(func(*http.Server) {}),
)
