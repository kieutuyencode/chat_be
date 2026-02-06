package http

import (
	"backend/conversation"
	"backend/file"
	"backend/http/handler"
	"backend/security/auth"
	"backend/user"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"go.uber.org/fx"
)

type routerParams struct {
	fx.In
	UserRouter         *user.Router
	FileRouter         *file.Router
	ConversationRouter *conversation.Router

	RequestTracking handler.RequestTracking
	ErrorHandler    handler.ErrorHandler
	VerifyUser      auth.VerifyUser
}

func newRouter(p routerParams) *iris.Application {
	router := iris.New()

	{
		router.Use(p.ErrorHandler)
		router.Configure(iris.WithRemoteAddrHeader(
			"X-Real-Ip",
			"X-Forwarded-For",
			"CF-Connecting-IP",
			"True-Client-Ip",
			"X-Appengine-Remote-Addr",
		))
		router.UseRouter(cors.AllowAll())
	}

	apiRouter := router.Party("/api")

	{
		apiRouter.Use(iris.Handler(p.RequestTracking))
		apiRouter.Use(iris.Handler(p.VerifyUser))
	}

	{
		v1Router := apiRouter.Party("/v1")

		p.UserRouter.Register(v1Router)
		p.FileRouter.Register(v1Router)
		p.ConversationRouter.Register(v1Router)
	}

	return router
}
