package graph

import "encoding/json"

func marshalTags(tags []string) string {
	if tags == nil {
		return "[]"
	}
	b, _ := json.Marshal(tags)
	return string(b)
}

func unmarshalTags(s string) []string {
	var tags []string
	_ = json.Unmarshal([]byte(s), &tags)
	if tags == nil {
		tags = []string{}
	}
	return tags
}

func marshalAttrs(attrs map[string]any) string {
	if attrs == nil {
		return "{}"
	}
	b, _ := json.Marshal(attrs)
	return string(b)
}

func unmarshalAttrs(s string) map[string]any {
	var attrs map[string]any
	_ = json.Unmarshal([]byte(s), &attrs)
	if attrs == nil {
		attrs = map[string]any{}
	}
	return attrs
}
