package log

type Config struct {
	// TimeFieldFormat defines the time format of the Time field type.
	TimeFieldFormat string
	// CallerSkipFrameCount is the number of stack frames to skip to find the caller.
	CallerSkipFrameCount *int
	// WithCaller enables adding the file:line of the caller.
	WithCaller bool
	// WithStack enables stack trace printing for the error.
	WithStack bool
}
