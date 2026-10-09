package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// YAMLFormatter formats data as clean YAML.
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAMLFormatter.
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format writes the YAML representation of data to w.
func (f *YAMLFormatter) Format(w io.Writer, data interface{}) error {
	rawJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var parsed interface{}
	if err := json.Unmarshal(rawJSON, &parsed); err != nil {
		return err
	}

	dumpYAML(w, parsed, 0)
	return nil
}

func dumpYAML(w io.Writer, val interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch v := val.(type) {
	case map[string]interface{}:
		for k, child := range v {
			switch child.(type) {
			case map[string]interface{}, []interface{}:
				fmt.Fprintf(w, "%s%s:\n", prefix, k)
				dumpYAML(w, child, indent+1)
			default:
				fmt.Fprintf(w, "%s%s: %v\n", prefix, k, formatYAMLValue(child))
			}
		}
	case []interface{}:
		for _, item := range v {
			switch child := item.(type) {
			case map[string]interface{}:
				fmt.Fprintf(w, "%s-\n", prefix)
				dumpYAML(w, child, indent+1)
			default:
				fmt.Fprintf(w, "%s- %v\n", prefix, formatYAMLValue(child))
			}
		}
	default:
		fmt.Fprintf(w, "%s%v\n", prefix, formatYAMLValue(v))
	}
}

func formatYAMLValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		if strings.Contains(val, "\n") || strings.Contains(val, ":") {
			return fmt.Sprintf("%q", val)
		}
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
