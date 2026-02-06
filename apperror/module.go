package apperror

import "go.uber.org/fx"

var Module = fx.Module("apperror",
	fx.Provide(newHandler),
)
