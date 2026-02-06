package file

import (
	"backend/apperror"
	"backend/common/result"
	"backend/http/validation"
	"backend/security/auth"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"go.uber.org/fx"
)

type Router struct {
	file File
}

type routerParams struct {
	fx.In
	File File
}

func newRouter(p routerParams) *Router {
	return &Router{
		file: p.File,
	}
}

func (r *Router) Register(routerGroup router.Party) {
	{
		router := routerGroup.Party("/file")

		router.Get("/client/{folderName}/{fileName}", validation.Validate[getClientFileParams](validation.ReadParams), func(ctx iris.Context) {
			params := ctx.Values().Get(string(validation.ReadParams)).(*getClientFileParams)
			subPath := params.FolderName
			cleanedDest, err := r.file.GetFilePath(subPath, params.FileName)

			if err != nil {
				ctx.SetErr(err)
				return
			}

			ctx.ServeFile(cleanedDest)
		})

		{
			requireUserRouter := router.Party("/", auth.RequireUser)

			requireUserRouter.Post("/upload", func(ctx iris.Context) {
				files, infos, err := ctx.FormFiles("files")
				if err != nil {
					ctx.SetErr(apperror.BadRequest("Upload file failed", nil, err))
					return
				}

				var data []map[string]string
				for i, file := range files {
					defer file.Close()

					src, err := r.file.SaveToTemporary(file, infos[i])
					if err != nil {
						ctx.SetErr(err)
						return
					}

					data = append(data, map[string]string{
						"src":  src,
						"name": infos[i].Filename,
					})
				}

				ctx.JSON(result.Success("Upload file success", data))
			})
		}
	}
}

type getClientFileParams struct {
	FolderName string `params:"folderName" validate:"required"`
	FileName   string `params:"fileName" validate:"required"`
}
