package providers

import (
	"bytes"
	"encoding/json"
	"strings"
)

// parseObject decodes one JSON object line; ok is false for anything else.
func parseObject(line []byte) (map[string]any, bool) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] != '{' {
		return nil, false
	}
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return nil, false
	}
	return m, true
}

// str returns the first string value among keys.
func str(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok {
			return s
		}
	}
	return ""
}

// num returns the first numeric value among keys.
func num(m map[string]any, keys ...string) (float64, bool) {
	for _, k := range keys {
		switch v := m[k].(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case json.Number:
			f, err := v.Float64()
			if err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

// inum is num truncated to int.
func inum(m map[string]any, keys ...string) int {
	f, _ := num(m, keys...)
	return int(f)
}

// sub returns a nested object or nil.
func sub(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	if s, ok := m[key].(map[string]any); ok {
		return s
	}
	return nil
}

// list returns a nested array or nil.
func list(m map[string]any, key string) []any {
	if m == nil {
		return nil
	}
	if l, ok := m[key].([]any); ok {
		return l
	}
	return nil
}

// findObject walks v depth-first and returns the first object for which
// pred is true.
func findObject(v any, pred func(map[string]any) bool) map[string]any {
	switch x := v.(type) {
	case map[string]any:
		if pred(x) {
			return x
		}
		for _, k := range sortedKeysAny(x) {
			if r := findObject(x[k], pred); r != nil {
				return r
			}
		}
	case []any:
		for _, e := range x {
			if r := findObject(e, pred); r != nil {
				return r
			}
		}
	}
	return nil
}

// sumNumbers walks v and sums every numeric value stored under one of the
// names (used for gemini stats whose shape is unverified).
func sumNumbers(v any, names ...string) (float64, bool) {
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}
	var total float64
	found := false
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, e := range x {
				if f, ok := e.(float64); ok && want[k] {
					total += f
					found = true
					continue
				}
				walk(e)
			}
		case []any:
			for _, e := range x {
				walk(e)
			}
		}
	}
	walk(v)
	return total, found
}

// hasString reports whether v contains s as a string value or key, anywhere.
func hasString(v any, s string) bool {
	switch x := v.(type) {
	case string:
		return x == s
	case map[string]any:
		for k, e := range x {
			if k == s || hasString(e, s) {
				return true
			}
		}
	case []any:
		for _, e := range x {
			if hasString(e, s) {
				return true
			}
		}
	}
	return false
}

func sortedKeysAny(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// lastObjectLine returns the last line that decodes as an object for which
// pred is true.
func lastObjectLine(stdout []byte, pred func(map[string]any) bool) map[string]any {
	var last map[string]any
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		m, ok := parseObject(line)
		if ok && pred(m) {
			last = m
		}
	}
	return last
}

// truncate shortens s for Message fields.
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// jsonMarshal renders v as indented JSON with a trailing newline.
func jsonMarshal(v any) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return append(b, '\n')
}
