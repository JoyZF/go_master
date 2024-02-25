package demo

import (
	"log/slog"
	"os"
	"testing"
)

func TestDemo1(t *testing.T) {
	opts := slog.HandlerOptions{
		AddSource: false,
	}
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &opts),
	)
	logger.Info("this is a message")
}
