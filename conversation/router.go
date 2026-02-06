package conversation

import (
	"backend/common/result"
	"backend/database"
	"backend/database/ent"
	"backend/http/pagination"
	"backend/http/validation"
	"backend/security/auth"
	"backend/security/jwt"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"go.uber.org/fx"
)

type Router struct {
	client       *ent.Client
	conversation *Conversation
}

type routerParams struct {
	fx.In
	Client       *ent.Client
	Conversation *Conversation
}

func newRouter(p routerParams) *Router {
	return &Router{
		client:       p.Client,
		conversation: p.Conversation,
	}
}

func (r *Router) Register(routerGroup router.Party) {
	{
		router := routerGroup.Party("/conversation")

		{
			requireUserRouter := router.Party("/", auth.RequireUser)

			requireUserRouter.Get("/online-users", func(ctx iris.Context) {
				claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
				res, err := r.conversation.GetOnlineUsers(ctx, r.client, &GetOnlineUsersParams{
					UserId: claims.UserId,
				})

				if err != nil {
					ctx.SetErr(err)
					return
				}

				ctx.JSON(result.Success("", res))
			})

			requireUserRouter.Post("/load", validation.Validate[loadBody](validation.ReadBody), func(ctx iris.Context) {
				body := ctx.Values().Get(string(validation.ReadBody)).(*loadBody)
				claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
				err := database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
					res, err := r.conversation.Load(ctx, tx.Client(), &LoadParams{
						FromUserId: claims.UserId,
						ToUserId:   body.UserId,
					})

					if err != nil {
						return err
					}

					ctx.JSON(result.Success("", res))
					return nil
				})

				if err != nil {
					ctx.SetErr(err)
					return
				}
			})

			requireUserRouter.Get("/", validation.Validate[getQuery](validation.ReadQuery), func(ctx iris.Context) {
				query := ctx.Values().Get(string(validation.ReadQuery)).(*getQuery)
				claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
				res, err := r.conversation.Get(ctx, r.client, &GetParams{
					UserId: claims.UserId,
					Limit:  query.Limit,
					Page:   query.Page,
					Search: query.Search,
				})

				if err != nil {
					ctx.SetErr(err)
					return
				}

				ctx.JSON(result.Success("", res))
			})

			requireUserRouter.Get("/{conversationId}", validation.Validate[getOneParams](validation.ReadParams), func(ctx iris.Context) {
				params := ctx.Values().Get(string(validation.ReadParams)).(*getOneParams)
				claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
				res, err := r.conversation.GetOne(ctx, r.client, &GetOneParams{
					UserId:         claims.UserId,
					ConversationId: params.ConversationId,
				})

				if err != nil {
					ctx.SetErr(err)
					return
				}

				ctx.JSON(result.Success("", res))
			})

			requireUserRouter.Post("/{conversationId}/message",
				validation.Validate[createMessageParams](validation.ReadParams),
				validation.Validate[createMessageBody](validation.ReadBody),
				func(ctx iris.Context) {
					database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
						params := ctx.Values().Get(string(validation.ReadParams)).(*createMessageParams)
						body := ctx.Values().Get(string(validation.ReadBody)).(*createMessageBody)
						claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
						mediaList := make([]*CreateMedia, len(body.Media))
						for i, m := range body.Media {
							mediaList[i] = &CreateMedia{
								Src: m.Src,
							}
						}
						res, err := r.conversation.CreateMessage(ctx, tx.Client(), &CreateMessageParams{
							UserId:         claims.UserId,
							ConversationId: params.ConversationId,
							Content:        body.Content,
							Media:          mediaList,
						})

						if err != nil {
							ctx.SetErr(err)
							return err
						}

						ctx.JSON(result.Success("Create message success", res))
						return nil
					})
				})

			requireUserRouter.Get("/{conversationId}/message",
				validation.Validate[getMessageParams](validation.ReadParams),
				validation.Validate[getMessageQuery](validation.ReadQuery),
				func(ctx iris.Context) {
					params := ctx.Values().Get(string(validation.ReadParams)).(*getMessageParams)
					query := ctx.Values().Get(string(validation.ReadQuery)).(*getMessageQuery)
					claims := ctx.Values().Get(auth.KeyUserClaims).(*jwt.UserClaims)
					res, err := r.conversation.GetMessage(ctx, r.client, &GetMessageParams{
						UserId:         claims.UserId,
						ConversationId: params.ConversationId,
						Limit:          query.Limit,
						Page:           query.Page,
					})

					if err != nil {
						ctx.SetErr(err)
						return
					}

					ctx.JSON(result.Success("", res))
				})
		}
	}

}

type loadBody struct {
	UserId int `json:"userId" validate:"required"`
}

type getQuery struct {
	pagination.Query
	Search string `query:"search"`
}

type getOneParams struct {
	ConversationId int `param:"conversationId" validate:"required"`
}

type createMessageBody struct {
	Content string         `json:"content"`
	Media   []*createMedia `json:"media"`
}

type createMessageParams struct {
	ConversationId int `param:"conversationId" validate:"required"`
}

type createMedia struct {
	Src string `json:"src"`
}

type getMessageQuery struct {
	pagination.Query
}

type getMessageParams struct {
	ConversationId int `param:"conversationId" validate:"required"`
}
