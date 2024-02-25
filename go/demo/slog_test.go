// Package demo slog demo
package demo

import (
	"context"
	"log/slog"
	"os"
)

// 自定义日志级别

func ExampleCustomLogLevel() {
	// Exported constants from a custom logging package.
	const (
		LevelTrace     = slog.Level(-8)
		LevelDebug     = slog.LevelDebug
		LevelInfo      = slog.LevelInfo
		LevelNotice    = slog.Level(2)
		LevelWarning   = slog.LevelWarn
		LevelError     = slog.LevelError
		LevelEmergency = slog.Level(12)
	)

	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     LevelTrace,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time from the output for predictable test output.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			// Customize the name of the level key and the output string, including
			// custom level values.
			if a.Key == slog.LevelKey {
				// Rename the level key from "level" to "sev".
				a.Key = "sev"

				// Handle custom level values.
				level := a.Value.Any().(slog.Level)

				// This could also look up the name from a map or other structure, but
				// this demonstrates using a switch statement to rename levels. For
				// maximum performance, the string values should be constants, but this
				// example uses the raw strings for readability.
				switch {
				case level < LevelDebug:
					a.Value = slog.StringValue("TRACE")
				case level < LevelInfo:
					a.Value = slog.StringValue("DEBUG")
				case level < LevelNotice:
					a.Value = slog.StringValue("INFO")
				case level < LevelWarning:
					a.Value = slog.StringValue("NOTICE")
				case level < LevelError:
					a.Value = slog.StringValue("WARNING")
				case level < LevelEmergency:
					a.Value = slog.StringValue("ERROR")
				default:
					a.Value = slog.StringValue("EMERGENCY")
				}
			}

			return a
		},
	})

	logger := slog.New(h)
	ctx := context.Background()
	logger.Log(ctx, LevelEmergency, "missing pilots")
	logger.Error("failed to start engines", "err", "missing fuel")
	logger.Warn("falling back to default value")
	logger.Log(ctx, LevelNotice, "all systems are running")
	logger.Info("initiating launch")
	logger.Debug("starting background job")
	logger.Log(ctx, LevelTrace, "button clicked")
	// Output:
	//sev=EMERGENCY source=/Users/joy/workspace/gopath/src/go_master/go/demo/slog_test.go:71 msg="missing pilots"
	//sev=ERROR source=/Users/joy/workspace/gopath/src/go_master/go/demo/slog_test.go:72 msg="failed to start engines" err="missing fuel"
	//sev=WARNING source=/Users/joy/workspace/gopath/src/go_master/go/demo/slog_test.go:73 msg="falling back to default value"
	//sev=NOTICE source=/Users/joy/workspace/gopath/src/go_master/go/demo/slog_test.go:74 msg="all systems are running"
	//sev=INFO source=/Users/joy/workspace/gopath/src/go_master/go/demo/slog_test.go:75 msg="initiating launch"
	//sev=DEBUG source=/Users/joy/workspace/gopath/src/go_master/go/demo/slog_test.go:76 msg="starting background job"
	//sev=TRACE source=/Users/joy/workspace/gopath/src/go_master/go/demo/slog_test.go:77 msg="button clicked"
}

// A LevelHandler wraps a Handler with an Enabled method
// that returns false for levels below a minimum.
type LevelHandler struct {
	level   slog.Leveler
	handler slog.Handler
}

// NewLevelHandler returns a LevelHandler with the given level.
// All methods except Enabled delegate to h.
func NewLevelHandler(level slog.Leveler, h slog.Handler) *LevelHandler {
	// Optimization: avoid chains of LevelHandlers.
	if lh, ok := h.(*LevelHandler); ok {
		h = lh.Handler()
	}
	return &LevelHandler{level, h}
}

// Enabled implements Handler.Enabled by reporting whether
// level is at least as large as h's level.
func (h *LevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements Handler.Handle.
func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithAttrs(attrs))
}

// WithGroup implements Handler.WithGroup.
func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithGroup(name))
}

// Handler returns the Handler wrapped by h.
func (h *LevelHandler) Handler() slog.Handler {
	return h.handler
}

// RemoveTime removes the top-level time attribute.
// It is intended to be used as a ReplaceAttr function,
// to make example output deterministic.
func RemoveTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

func ExampleHandler_levelHandler() {
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: RemoveTime})
	logger := slog.New(NewLevelHandler(slog.LevelWarn, th))
	logger.Info("not printed")
	logger.Warn("printed")

	// Output:
	// level=WARN msg=printed
}

type Name struct {
	First, Last string
}

// LogValue implements slog.LogValuer.
// It returns a group containing the fields of
// the Name, so that they appear together in the log output.
func (n Name) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("first", n.First),
		slog.String("last", n.Last))
}

func ExampleLogValuer_group() {
	n := Name{"Perry", "Platypus"}
	th := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: RemoveTime,
	})
	logger := slog.New(th)
	logger.Info("mission accomplished", "agent", n)

	//Output:
	// {"level":"INFO","msg":"mission accomplished","agent":{"first":"Perry","last":"Platypus"}}
}

// A token is a secret value that grants permissions.
type Token string

// LogValue implements slog.LogValuer.
// It avoids revealing the token.
func (Token) LogValue() slog.Value {
	return slog.StringValue("REDACTED_TOKEN")
}

// This example demonstrates a Value that replaces itself
// with an alternative representation to avoid revealing secrets.
func ExampleLogValuer_secret() {
	t := Token("shhhh!")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: RemoveTime}))
	logger.Info("permission granted", "user", "Perry", "token", t)

	// Output:
	// level=INFO msg="permission granted" user=Perry token=REDACTED_TOKEN
}
