package packet

import (
	"encoding/json"
	"strings"
	"testing"
)

// The unit schema must be valid in OpenAI strict mode: every object
// closed, every property required. It offers no_information and the three
// support levels, and no plan.
func TestUnitSchemaIsStrictAndOffersNoInformation(t *testing.T) {
	var s map[string]any
	if err := json.Unmarshal([]byte(UnitSchema()), &s); err != nil {
		t.Fatal(err)
	}
	var walk func(path string, v map[string]any)
	walk = func(path string, v map[string]any) {
		if v["type"] == "object" {
			props, _ := v["properties"].(map[string]any)
			req, _ := v["required"].([]any)
			if v["additionalProperties"] != false || len(req) != len(props) {
				t.Errorf("%s: %d properties, %d required, additionalProperties %v", path, len(props), len(req), v["additionalProperties"])
			}
			for k, p := range props {
				walk(path+"."+k, p.(map[string]any))
			}
		}
		if items, ok := v["items"].(map[string]any); ok {
			walk(path+"[]", items)
		}
	}
	walk("unit", s)
	schema := UnitSchema()
	for _, want := range []string{`"no_information"`, `"quoted"`, `"inferred"`, `"unsupported"`} {
		if !strings.Contains(schema, want) {
			t.Errorf("schema lacks %s", want)
		}
	}
	if strings.Contains(schema, `"plan"`) || strings.Contains(schema, `"units"`) {
		t.Error("a unit cannot plan")
	}
}

func TestParseUnitHandoff(t *testing.T) {
	cases := []struct {
		name, json string
		ok         bool
		claims     int
	}{
		{"answer", `{"status":"completed","answer":"Done."}`, true, 0},
		{"claims only", `{"status":"completed","answer":"","claims":[{"claim":"The port is 5433.","evidence":["t000004"],"quote":"port 5433","support":"quoted"}]}`, true, 1},
		{"nothing", `{"status":"completed","answer":"  ","claims":[{"claim":" ","evidence":[],"quote":"","support":"quoted"}]}`, false, 0},
		{"no information", `{"status":"no_information","answer":"","notes":"Looked for a budget; none stated."}`, true, 0},
		{"blocked empty", `{"status":"blocked","answer":""}`, false, 0},
		{"recall", `{"status":"needs_recall","recall_ids":["t000002"]}`, true, 0},
		{"recall empty", `{"status":"needs_recall"}`, false, 0},
		{"plan", `{"status":"plan","units":[{"id":"u1"}]}`, false, 0},
		{"partial", `{"status":"partial","answer":"first half","next_step":"second half"}`, true, 0},
	}
	for _, c := range cases {
		h, err := ParseUnitHandoff([]byte(c.json))
		if (err == nil) != c.ok {
			t.Errorf("%s: err %v, want ok %v", c.name, err, c.ok)
			continue
		}
		if err == nil && len(h.Claims) != c.claims {
			t.Errorf("%s: %d claims, want %d", c.name, len(h.Claims), c.claims)
		}
	}
	// a respond handoff never takes no_information
	if _, err := ParseHandoff([]byte(`{"status":"no_information","answer":""}`)); err == nil {
		t.Error("ParseHandoff accepted no_information")
	}
}
