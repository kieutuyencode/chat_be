package notification

import (
	"backend/notification/mail"

	"go.uber.org/fx"
)

var Module = fx.Module("notification",
	mail.Module,
)
