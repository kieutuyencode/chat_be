package user

import (
	"backend/user/auth"
	"backend/user/profile"

	"github.com/kataras/iris/v12/core/router"
	"go.uber.org/fx"
)

type Router struct {
	authRouter    *auth.Router
	profileRouter *profile.Router
}

type routerParams struct {
	fx.In
	AuthRouter    *auth.Router
	ProfileRouter *profile.Router
}

func newRouter(p routerParams) *Router {
	return &Router{
		authRouter:    p.AuthRouter,
		profileRouter: p.ProfileRouter,
	}
}

func (r *Router) Register(routerGroup router.Party) {
	{
		router := routerGroup.Party("/user")

		r.authRouter.Register(router)
		r.profileRouter.Register(router)
	}
}
