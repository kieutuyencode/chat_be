package security

import (
	"backend/security/auth"
	"backend/security/jwt"

	"go.uber.org/fx"
)

var Module = fx.Module("security",
	jwt.Module,
	auth.Module,
)
