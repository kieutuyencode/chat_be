package auth

import (
	"backend/apperror"
	"backend/http/key"
	"backend/security/jwt"
	"strconv"

	"github.com/kataras/iris/v12"
	"go.uber.org/fx"
)

type VerifyUser iris.Handler

type verifyUserParams struct {
	fx.In
	Jwt jwt.Jwt
}

func newVerifyUser(p verifyUserParams) VerifyUser {
	return func(ctx iris.Context) {
		accessToken := ctx.Request().Header.Get(keyAccessTokenHeader)

		if accessToken == "" {
			ctx.Next()
			return
		}

		claims := &jwt.UserClaims{}
		err := p.Jwt.ValidateAccessToken(accessToken, claims)
		if err != nil {
			ctx.SetErr(apperror.Unauthorized(messageInvalidOrExpiredToken, nil, err))
			return
		}

		ctx.Values().Set(KeyUserClaims, claims)
		ctx.Values().Set(key.ClientId, strconv.Itoa(claims.UserId))

		ctx.Next()
	}
}
