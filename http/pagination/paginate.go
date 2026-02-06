package pagination

import (
	"context"
	"database/sql"
	"errors"
)

type Result[T any] struct {
	Count int `json:"count"`
	Rows  []T `json:"rows"`
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

func Paginate[T any, Q interface {
	Limit(int) Q
	Offset(int) Q
	All(context.Context) ([]T, error)
	Count(context.Context) (int, error)
}](ctx context.Context, queryEnt Q, query *Query) (*Result[T], error) {
	limit, offset := query.LimitOffset()

	total, err := queryEnt.Count(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	rows, err := queryEnt.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, err
	}

	return &Result[T]{
		Count: total,
		Rows:  rows,
		Limit: limit,
		Page:  query.Page,
	}, nil
}
