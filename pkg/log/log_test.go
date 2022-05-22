package log

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestInitWithConfig(t *testing.T) {
	out := &bytes.Buffer{}

	callerSkipFrameCount := 3
	Init(out, &Config{
		TimeFieldFormat:      time.RFC822,
		CallerSkipFrameCount: &callerSkipFrameCount,
		WithCaller:           true,
		WithStack:            true,
	})

	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2001, time.February, 3, 4, 5, 6, 7, time.UTC)
	}

	_, file, line, _ := runtime.Caller(0)
	caller := callerMarshal(file, line+3)

	Info(context.Background(), "hello, world", "foo", "bar")

	got := out.String()
	want := fmt.Sprintf(`{"level":"info","foo":"bar","time":"03 Feb 01 04:05 UTC","caller":"%s","message":"hello, world"}`+"\n", caller)

	if got != want {
		t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
	}
}

func TestGlobalLevel(t *testing.T) {
	out := &bytes.Buffer{}

	Init(out, &Config{
		WithCaller: false,
		WithStack:  false,
	})

	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2001, time.February, 3, 4, 5, 6, 7, time.UTC)
	}

	SetGlobalLevel(ErrorLevel)

	Debug(context.Background(), "this message should not be printed")
	if out.String() != "" {
		t.Errorf("expected empty log output,\ngot: %v", out.String())
	}

	Error(context.Background(), "got fatal error", errors.New("fatal"))
	got := out.String()
	want := fmt.Sprint(`{"level":"error","error":"fatal","time":"2001-02-03T04:05:06.000000007Z","message":"got fatal error"}` + "\n")
	if got != want {
		t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
	}
}

func TestWithValues(t *testing.T) {
	SetGlobalLevel(DebugLevel)
	out := &bytes.Buffer{}

	Init(out, &Config{
		WithCaller: false,
		WithStack:  false,
	})

	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2001, time.February, 3, 4, 5, 6, 7, time.UTC)
	}

	lg := WithValues(context.Background(), "some value", "test")

	lg.Error(errors.New("fatal"), "got fatal error")
	got := out.String()
	want := fmt.Sprint(`{"level":"error","some value":"test","error":"fatal","time":"2001-02-03T04:05:06.000000007Z","message":"got fatal error"}` + "\n")
	if got != want {
		t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
	}
	out.Reset()

	lg.Info("some information")
	got = out.String()
	want = fmt.Sprint(`{"level":"debug","some value":"test","time":"2001-02-03T04:05:06.000000007Z","message":"some information"}` + "\n")
	if got != want {
		t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
	}
}
