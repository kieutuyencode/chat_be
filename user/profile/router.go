package profile

import (
	"backend/common/result"
	"backend/database/ent"
	"backend/http/validation"
	"backend/security/auth"
	"backend/security/jwt"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"go.uber.org/fx"
)

type Router struct {
	client  *ent.Client
	profile *Profile
}

type routerParams struct {
	fx.In
	Client  *ent.Client
	Profile *Profile
}

func newRouter(p routerParams) *Router {
	return &Router{
		client:  p.Client,
		profile: p.Profile,
	}
}

func (r *Router) Register(routerGroup router.Party) {
	{
		router := routerGroup.Party("/profile")

		requireUserRouter := router.Party("/", auth.RequireUser)

		requireUserRouter.Get("/", func(ctx iris.Context) {
			claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
			res, err := r.profile.GetProfile(ctx, r.client, claims.UserId)

			if err != nil {
				ctx.SetErr(err)
				return
			}

			ctx.JSON(result.Success("", res))
		})

		requireUserRouter.Patch("/", validation.Validate[updateProfileBody](validation.ReadBody), func(ctx iris.Context) {
			claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
			body := ctx.Values().Get(string(validation.ReadBody)).(*updateProfileBody)
			res, err := r.profile.UpdateProfile(ctx, r.client, &UpdateProfileParams{
				UserId:   claims.UserId,
				Fullname: body.Fullname,
				Phone:    body.Phone,
				Avatar:   body.Avatar,
			})

			if err != nil {
				ctx.SetErr(err)
				return
			}

			ctx.JSON(result.Success("Update success", res))
		})
	}

}

type updateProfileBody struct {
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}
