package websocket

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/philippseith/signalr"
	"go.uber.org/fx"
)

type Server struct {
	signalr.Server
}

type serverParams struct {
	fx.In
	Websocket *Websocket
}

func newServer(p serverParams) (*Server, error) {
	srv, err := signalr.NewServer(context.TODO(),
		signalr.UseHub(p.Websocket),
		signalr.InsecureSkipVerify(true),
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create signalr server")
	}

	return &Server{Server: srv}, nil
}
