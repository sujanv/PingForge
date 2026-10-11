package output

import (
	"encoding/json"
	"io"

	"github.com/sujanv/pingforge/internal/core"
)

// NDJSONFormatter outputs Newline Delimited JSON.
type NDJSONFormatter struct{}

// NewNDJSONFormatter creates a new NDJSONFormatter.
func NewNDJSONFormatter() *NDJSONFormatter {
	return &NDJSONFormatter{}
}

// Format streams each item on a separate newline.
func (n *NDJSONFormatter) Format(w io.Writer, data interface{}) error {
	switch v := data.(type) {
	case []core.ProbeResult:
		for _, item := range v {
			b, err := json.Marshal(item)
			if err != nil {
				return err
			}
			if _, err := w.Write(append(b, '\n')); err != nil {
				return err
			}
		}
	default:
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		_, err = w.Write(append(b, '\n'))
		return err
	}
	return nil
}
