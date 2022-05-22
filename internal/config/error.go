package config

import (
	"fmt"
)

// ErrMissingOption is the error when configuration option is required
// in a particular situation but not provided by the user.
type ErrMissingOption struct {
	Option string
}

func (e ErrMissingOption) Error() string {
	return fmt.Sprintf("Missing configuration option [%s]", e.Option)
}
