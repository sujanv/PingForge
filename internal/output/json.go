package output

import (
	"encoding/json"
	"io"
)

// JSONFormatter serializes results as indented JSON.
type JSONFormatter struct {
	Indent bool
}

// NewJSONFormatter creates a new JSONFormatter.
func NewJSONFormatter(indent bool) *JSONFormatter {
	return &JSONFormatter{Indent: indent}
}

// Format writes the JSON representation of data to w.
func (f *JSONFormatter) Format(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	if f.Indent {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(data)
}
