package providers

import (
	"encoding/json"
	"testing"
)

func TestStrictSchema(t *testing.T) {
	in := `{"type":"object","properties":{"status":{"type":"string"},"units":{"type":"array","items":{"type":"object","properties":{"id":{"type":"string"},"n":{"type":"integer"}}}}},"required":["status"]}`
	out, err := StrictSchema([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if m["additionalProperties"] != false {
		t.Error("top level additionalProperties")
	}
	req, _ := m["required"].([]any)
	if len(req) != 2 || req[0] != "status" || req[1] != "units" {
		t.Errorf("required %v", req)
	}
	items := m["properties"].(map[string]any)["units"].(map[string]any)["items"].(map[string]any)
	if items["additionalProperties"] != false || len(items["required"].([]any)) != 2 {
		t.Errorf("nested items %v", items)
	}
	if _, err := StrictSchema([]byte("nope")); err == nil {
		t.Error("invalid json must fail")
	}
}

func TestStrictProblems(t *testing.T) {
	ok := []byte(`{"type":"object","properties":{"a":{"type":["string","null"]},"b":{"type":"array","items":{"type":"object","properties":{"c":{"type":"integer"}},"required":["c"],"additionalProperties":false}}},"required":["a","b"],"additionalProperties":false}`)
	if p := StrictProblems(ok); len(p) != 0 {
		t.Fatalf("strict schema flagged: %v", p)
	}
	cases := map[string]string{
		"top-level allOf": `{"type":"object","properties":{},"required":[],"additionalProperties":false,"allOf":[{}]}`,
		"open object":     `{"type":"object","properties":{"m":{"type":"object"}},"required":["m"],"additionalProperties":false}`,
		"optional key":    `{"type":"object","properties":{"a":{"type":"string"}},"additionalProperties":false}`,
		"no type":         `{"type":"object","properties":{"a":{}},"required":["a"],"additionalProperties":false}`,
		"const":           `{"type":"object","properties":{"a":{"type":"string","const":"x"}},"required":["a"],"additionalProperties":false}`,
	}
	for name, s := range cases {
		if p := StrictProblems([]byte(s)); len(p) == 0 {
			t.Errorf("%s: not flagged", name)
		}
	}
}
