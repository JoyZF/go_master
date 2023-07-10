package impl

import (
	"context"
	"errors"
)

type Rose struct {
}

func New() *Rose {
	return &Rose{}
}

func (r *Rose) Init(ctx context.Context, size int64) error {
	if int64(int(size)) != size {
		return errors.New("size error")
	}
	return nil
}

func (r *Rose) Set(ctx context.Context, key string, value string) error {
	return nil
}

func (r *Rose) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}
