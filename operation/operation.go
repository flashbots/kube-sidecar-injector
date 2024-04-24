package operation

import (
	"encoding/json"
	"strings"

	json_patch "github.com/evanphx/json-patch"
)

var (
	rawAdd     = json.RawMessage(`"add"`)
	rawReplace = json.RawMessage(`"replace"`)
)

func Escape(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

func Add(path string, value interface{}) (
	json_patch.Operation, error,
) {
	bytesPath, err := json.Marshal(path)
	if err != nil {
		return nil, err
	}
	rawPath := json.RawMessage(bytesPath)

	bytesValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	rawValue := json.RawMessage(bytesValue)

	return map[string]*json.RawMessage{
		"op":    &rawAdd,
		"path":  &rawPath,
		"value": &rawValue,
	}, nil
}

func Replace(path string, value interface{}) (
	json_patch.Operation, error,
) {
	bytesPath, err := json.Marshal(path)
	if err != nil {
		return nil, err
	}
	rawPath := json.RawMessage(bytesPath)

	bytesValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	rawValue := json.RawMessage(bytesValue)

	return map[string]*json.RawMessage{
		"op":    &rawReplace,
		"path":  &rawPath,
		"value": &rawValue,
	}, nil
}
