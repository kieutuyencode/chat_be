package user

import (
	"backend/user/auth"
	"backend/user/profile"

	"go.uber.org/fx"
)

var Module = fx.Module("user",
	fx.Provide(newRouter),
	auth.Module,
	profile.Module,
)
