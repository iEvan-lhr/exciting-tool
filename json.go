package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func jsonInput(v any) ([]byte, error) {
	switch value := v.(type) {
	case string:
		return []byte(value), nil
	case []byte:
		return value, nil
	default:
		return json.Marshal(value)
	}
}

// Unmarshal parses JSON from a string, byte slice, or marshalable Go value.
// Deprecated: Use encoding/json.Unmarshal to receive an error.
func Unmarshal(v interface{}, target interface{}) {
	data, err := jsonInput(v)
	ExecError(err)
	if target == nil {
		return
	}
	ExecError(json.Unmarshal(data, target))
}

// UnmarshalByOriginal is retained for compatibility and uses encoding/json
// without post-processing the source.
// Deprecated: Use encoding/json.Unmarshal.
func UnmarshalByOriginal(v interface{}, target interface{}) {
	Unmarshal(v, target)
}

// UMarshal is retained for compatibility with the original API.
// Deprecated: Use encoding/json.Unmarshal.
func UMarshal(v, target interface{}) {
	Unmarshal(v, target)
}

// Marshal serializes a value while leaving HTML characters unescaped.
// Deprecated: Use json.Encoder with SetEscapeHTML(false) to receive an error.
func Marshal(v interface{}) []byte {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	ExecError(encoder.Encode(v))
	return bytes.TrimSuffix(buffer.Bytes(), []byte{'\n'})
}

// MarshalJSON makes String interoperate with encoding/json as a JSON string.
func (s *String) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	return json.Marshal(s.String())
}

// UnmarshalJSON accepts JSON strings as well as scalar numbers and booleans.
func (s *String) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		s.reset()
		return nil
	}

	var value string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
	} else {
		var scalar any
		if err := json.Unmarshal(data, &scalar); err != nil {
			return err
		}
		switch scalar.(type) {
		case float64, bool:
			value = string(data)
		default:
			return fmt.Errorf("tools.String cannot decode JSON value %s", data)
		}
	}

	s.reset()
	s.Append(value)
	return nil
}
