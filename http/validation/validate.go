package validation

import (
	"backend/apperror"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
)

var validate = validator.New()

func Validate[T any](readFuncName read, opts ...option) iris.Handler {
	cfg := &config{
		validate,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx iris.Context) {
		var data T
		var err error

		switch readFuncName {
		case ReadQuery:
			err = ctx.ReadQuery(&data)
		case ReadParams:
			err = ctx.ReadParams(&data)
		case ReadHeaders:
			err = ctx.ReadHeaders(&data)
		case ReadBody:
			err = ctx.ReadJSON(&data)
		}

		if err != nil {
			ctx.SetErr(apperror.BadRequest("Bind data failed", nil, err))
			return
		}

		if err = cfg.validate.Struct(data); err != nil {
			ctx.SetErr(apperror.BadRequest("Validation failed", nil, err))
			return
		}

		ctx.Values().Set(string(readFuncName), &data)
		ctx.Next()
	}
}
