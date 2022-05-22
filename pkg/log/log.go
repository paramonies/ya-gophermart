package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"

	"github.com/paramonies/ya-gophermart/pkg/log/requestid"
	"github.com/paramonies/ya-gophermart/pkg/log/zerologr"
)

const (
	defaultTimeFieldFormat      = time.RFC3339Nano
	defaultCallerSkipFrameCount = 3
)

const (
	requestIDKey = "x-request-id"
)

var Logger zerologr.Logger

func init() {
	zl := zerolog.Nop()
	Logger = zerologr.New(&zl)
}

func InitDefault() {
	zerolog.TimeFieldFormat = defaultTimeFieldFormat
	zerolog.CallerSkipFrameCount = defaultCallerSkipFrameCount
	zerolog.CallerMarshalFunc = callerMarshal
	zerolog.LevelWarnValue = "warning"
	zl := zerolog.New(os.Stderr).With().Caller().Stack().Timestamp().Logger()

	Logger = zerologr.New(&zl)
}

func Init(w io.Writer, cfg *Config) {
	zerolog.LevelWarnValue = "warning"

	if cfg.TimeFieldFormat != "" {
		zerolog.TimeFieldFormat = cfg.TimeFieldFormat
	} else {
		zerolog.TimeFieldFormat = defaultTimeFieldFormat
	}

	if cfg.CallerSkipFrameCount != nil {
		zerolog.CallerSkipFrameCount = *cfg.CallerSkipFrameCount
	} else {
		zerolog.CallerSkipFrameCount = defaultCallerSkipFrameCount
	}

	zerolog.CallerMarshalFunc = callerMarshal

	zlCtx := zerolog.New(w).With().Timestamp()
	if cfg.WithCaller {
		zlCtx = zlCtx.Caller()
	}
	if cfg.WithStack {
		zlCtx = zlCtx.Stack()
	}

	zl := zlCtx.Logger()
	Logger = zerologr.New(&zl)
}

func SetGlobalLevel(l Level) {
	zerolog.SetGlobalLevel(l)
}

func Debug(ctx context.Context, msg string, kv ...interface{}) {
	kv = withRequestID(ctx, kv)
	Logger.V(0).Info(msg, kv...)
}

func Info(ctx context.Context, msg string, kv ...interface{}) {
	kv = withRequestID(ctx, kv)
	Logger.V(1).Info(msg, kv...)
}

func Warning(ctx context.Context, msg string, kv ...interface{}) {
	kv = withRequestID(ctx, kv)
	Logger.V(2).Info(msg, kv...)
}

func Error(ctx context.Context, msg string, err error, kv ...interface{}) {
	kv = withRequestID(ctx, kv)
	Logger.Error(err, msg, kv...)
}

func WithValues(ctx context.Context, kv ...interface{}) zerologr.Logger {
	kv = withRequestID(ctx, kv)
	return Logger.WithValues(kv...)
}

func withRequestID(ctx context.Context, kv []interface{}) []interface{} {
	rid := requestid.FromContext(ctx)
	if rid == "" {
		return kv
	}

	return append(kv, requestIDKey, rid)
}

func callerMarshal(file string, line int) string {
	base := filepath.Base(file)
	return fmt.Sprintf("%s:%d", base, line)
}
