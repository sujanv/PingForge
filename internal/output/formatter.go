package output

import (
	"io"
)

// Formatter defines formatting behavior for CLI results.
type Formatter interface {
	Format(w io.Writer, data interface{}) error
}
