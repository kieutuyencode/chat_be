package database

import (
	"backend/config"
	"backend/database/ent"
	"context"

	"entgo.io/ent/dialect"
	"github.com/cockroachdb/errors"
	"go.uber.org/fx"
)

func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = errors.Newf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Newf("committing transaction: %w", err)
	}

	return nil
}

type clientParams struct {
	fx.In
	fx.Lifecycle
	Env *config.Env
}

func newClient(p clientParams) (*ent.Client, error) {
	client, err := ent.Open(dialect.Postgres, p.Env.DBUrl)

	if err != nil {
		return nil, errors.Wrapf(err, "failed opening connection to %s", dialect.Postgres)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			defer client.Close()
			return nil
		},
	})

	return client, nil
}
