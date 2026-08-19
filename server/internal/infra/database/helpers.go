package database

import "encoding/json"

func marshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func unmarshalJSON[T any](s string, target *T) {
	if s == "" || s == "null" {
		return
	}
	_ = json.Unmarshal([]byte(s), target)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func boolPtr(b int) *bool {
	v := b == 1
	return &v
}
