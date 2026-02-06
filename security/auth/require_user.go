package auth

import (
	"backend/apperror"

	"github.com/kataras/iris/v12"
)

func RequireUser(ctx iris.Context) {
	claims := ctx.Values().Get(KeyUserClaims)
	if claims == nil {
		ctx.SetErr(apperror.Unauthorized("Unauthorized", nil, nil))
		return
	}

	ctx.Next()
}
