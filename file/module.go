package file

import "go.uber.org/fx"

var Module = fx.Module("file",
	fx.Provide(newRouter, newFile),
)
