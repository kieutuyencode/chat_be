package main

import (
	"backend/apperror"
	"backend/config"
	"backend/conversation"
	"backend/database"
	"backend/file"
	"backend/http"
	"backend/logger"
	"backend/notification"
	"backend/security"
	"backend/user"
	"backend/websocket"

	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		logger.Module,
		apperror.Module,
		config.Module,
		database.Module,
		notification.Module,
		websocket.Module,
		http.Module,
		security.Module,
		user.Module,
		file.Module,
		conversation.Module,
	).Run()
}
