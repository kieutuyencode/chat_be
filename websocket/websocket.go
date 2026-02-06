package websocket

import (
	"backend/apperror"
	"backend/common/result"
	"backend/database/ent"
	entuser "backend/database/ent/user"
	"backend/security/jwt"
	"context"
	"time"

	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/philippseith/signalr"
	"go.uber.org/fx"
)

type Websocket struct {
	jwt     jwt.Jwt
	handler apperror.Handler
	client  *ent.Client
	signalr.Hub
}

type websocketParams struct {
	fx.In
	fx.Lifecycle
	Jwt     jwt.Jwt
	Handler apperror.Handler
	Client  *ent.Client
}

func newWebsocket(p websocketParams) *Websocket {
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_, err := p.Client.User.Update().SetIsActive(false).Save(ctx)
			if err != nil {
				return errors.Wrap(err, "Update user failed")
			}
			return nil
		},
	})

	return &Websocket{
		jwt:     p.Jwt,
		handler: p.Handler,
		client:  p.Client,
	}
}

func (w *Websocket) OnConnected(connectionID string) {
}

func (w *Websocket) OnDisconnected(connectionID string) {
	w.handler(func() error {
		ctx := context.Background()

		userData, ok := w.Items().Load("user")
		if !ok {
			return nil
		}
		user, ok := userData.(*ent.User)
		if !ok {
			return nil
		}

		_, err := user.Update().SetIsActive(false).SetLastActiveAt(time.Now()).Save(ctx)
		if err != nil {
			return errors.Wrap(err, "Update user failed")
		}

		w.Items().Clear()
		w.Clients().All().Send(eventUserConnection, "")

		return nil
	})
}

func (w *Websocket) Connect(accessToken string) {
	w.handler(func() error {
		ctx := context.Background()
		target := "connect"
		claims := &jwt.UserClaims{}
		err := w.jwt.ValidateAccessToken(accessToken, claims)

		if err != nil {
			w.Clients().Caller().Send(target, result.Fail("Token invalid or expired", nil))
			return nil
		}

		user, err := w.client.User.Query().Where(entuser.ID(claims.UserId)).First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return errors.Wrap(err, "Get user failed")
		}
		if user == nil {
			w.Clients().Caller().Send(target, result.Fail("User not found", nil))
			return nil
		}

		user, err = user.Update().SetIsActive(true).SetLastActiveAt(time.Now()).Save(ctx)
		if err != nil {
			return errors.Wrap(err, "Update user failed")
		}

		w.Groups().AddToGroup(strconv.Itoa(int(claims.UserId)), w.ConnectionID())
		w.Clients().Caller().Send(target, result.Success("Connect success", user))

		w.Items().Store("user", user)
		w.Clients().All().Send(eventUserConnection, "")

		return nil
	})
}

const (
	eventUserConnection  = "userConnection"
	EventMessageReceived = "messageReceived"
	EventMessageSeen     = "messageSeen"
)
