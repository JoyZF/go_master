package _interface

import "context"



type Store interface {
    Init(ctx context.Context, size int64) error
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
}

