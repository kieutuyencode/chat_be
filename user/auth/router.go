package auth

import (
	"backend/common/result"
	"backend/database/ent"
	"backend/http/validation"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"go.uber.org/fx"
)

type Router struct {
	client *ent.Client
	auth   *Auth
}

type routerParams struct {
	fx.In
	Client *ent.Client
	Auth   *Auth
}

func newRouter(p routerParams) *Router {
	return &Router{
		client: p.Client,
		auth:   p.Auth,
	}
}

func (r *Router) Register(routerGroup router.Party) {
	{
		router := routerGroup.Party("/auth")

		router.Post("/sign-in", validation.Validate[signInBody](validation.ReadBody), func(ctx iris.Context) {
			body := ctx.Values().Get(string(validation.ReadBody)).(*signInBody)
			err := r.auth.SignIn(ctx, r.client, &SignInParams{
				Email: body.Email,
			})

			if err != nil {
				ctx.SetErr(err)
				return
			}

			ctx.JSON(result.Success("Vui lòng kiểm tra email để xác nhận đăng nhập", nil))
		})

		router.Post("/verify-sign-in", validation.Validate[verifySignInBody](validation.ReadBody), func(ctx iris.Context) {
			body := ctx.Values().Get(string(validation.ReadBody)).(*verifySignInBody)
			accessToken, err := r.auth.VerifySignIn(ctx, r.client, &VerifySignInParams{
				Email: body.Email,
				Code:  body.Code,
			})

			if err != nil {
				ctx.SetErr(err)
				return
			}

			ctx.JSON(result.Success("Sign in successfully", map[string]string{
				"accessToken": accessToken,
			}))
		})
	}

}

type signInBody struct {
	Email string `json:"email" validate:"required,email"`
}

type verifySignInBody struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}
