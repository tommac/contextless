package providers

import (
	"encoding/json"
	"fmt"
	"sort"
)

// StrictSchema converts a JSON schema to OpenAI strict form for codex
// --output-schema (04 §4.3): every object lists all of its properties as
// required and sets additionalProperties false, recursively through
// properties, items, prefixItems, anyOf, oneOf, allOf, $defs and
// definitions. Keys come out sorted.
func StrictSchema(schema []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(schema, &v); err != nil {
		return nil, fmt.Errorf("providers: strict schema: %w", err)
	}
	strictify(v)
	return json.MarshalIndent(v, "", "  ")
}

func strictify(v any) {
	switch x := v.(type) {
	case map[string]any:
		props, hasProps := x["properties"].(map[string]any)
		if x["type"] == "object" || hasProps {
			x["additionalProperties"] = false
			if hasProps {
				req := make([]string, 0, len(props))
				for k := range props {
					req = append(req, k)
				}
				sort.Strings(req)
				x["required"] = req
				for _, p := range props {
					strictify(p)
				}
			}
		}
		for _, key := range []string{"items", "additionalItems", "not"} {
			if c, ok := x[key]; ok {
				strictify(c)
			}
		}
		for _, key := range []string{"prefixItems", "anyOf", "oneOf", "allOf"} {
			if l, ok := x[key].([]any); ok {
				for _, c := range l {
					strictify(c)
				}
			}
		}
		for _, key := range []string{"$defs", "definitions"} {
			if d, ok := x[key].(map[string]any); ok {
				for _, c := range d {
					strictify(c)
				}
			}
		}
	case []any:
		for _, c := range x {
			strictify(c)
		}
	}
}

// strictBanned are keywords the OpenAI strict dialect rejects anywhere in an
// --output-schema document.
var strictBanned = []string{"allOf", "oneOf", "not", "if", "then", "else", "const",
	"dependentRequired", "dependentSchemas", "patternProperties", "unevaluatedProperties",
	"propertyNames", "minProperties", "maxProperties", "contains", "prefixItems"}

// StrictProblems lists the places where a schema breaks the OpenAI strict
// rules that codex --output-schema enforces with a 400 before the model
// runs: the root is a plain object (no anyOf, oneOf, allOf, enum or not at
// the top level); every subschema has a type, an enum, or an anyOf; every
// object lists its properties, sets additionalProperties false, and
// requires every property; none of strictBanned appears. An open object (a
// map keyed by ids) cannot be expressed, so a contract that has one cannot
// be strict. Empty means the schema is safe to pass.
func StrictProblems(schema []byte) []string {
	var v any
	if err := json.Unmarshal(schema, &v); err != nil {
		return []string{"not JSON: " + err.Error()}
	}
	var out []string
	root, ok := v.(map[string]any)
	if !ok {
		return []string{"$: not an object schema"}
	}
	if root["type"] != "object" {
		out = append(out, "$: the root type must be object")
	}
	for _, k := range []string{"anyOf", "enum"} {
		if _, has := root[k]; has {
			out = append(out, "$: "+k+" at the top level")
		}
	}
	strictWalk("$", root, &out)
	return out
}

func strictWalk(path string, v any, out *[]string) {
	m, ok := v.(map[string]any)
	if !ok {
		return
	}
	for _, b := range strictBanned {
		if _, has := m[b]; has {
			*out = append(*out, path+": "+b)
		}
	}
	_, hasType := m["type"]
	_, hasAny := m["anyOf"]
	_, hasEnum := m["enum"]
	if !hasType && !hasAny && !hasEnum {
		*out = append(*out, path+": no type")
	}
	isObj := m["type"] == "object"
	if l, ok := m["type"].([]any); ok {
		for _, t := range l {
			if t == "object" {
				isObj = true
			}
		}
	}
	if isObj {
		props, _ := m["properties"].(map[string]any)
		if props == nil {
			*out = append(*out, path+": open object (no properties)")
		}
		if m["additionalProperties"] != false {
			*out = append(*out, path+": additionalProperties is not false")
		}
		req := map[string]bool{}
		if l, ok := m["required"].([]any); ok {
			for _, r := range l {
				if s, ok := r.(string); ok {
					req[s] = true
				}
			}
		}
		keys := make([]string, 0, len(props))
		for k := range props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !req[k] {
				*out = append(*out, path+"."+k+": not required")
			}
			strictWalk(path+"."+k, props[k], out)
		}
	}
	if it, ok := m["items"]; ok {
		strictWalk(path+"[]", it, out)
	}
	if l, ok := m["anyOf"].([]any); ok {
		for i, c := range l {
			strictWalk(fmt.Sprintf("%s|%d", path, i), c, out)
		}
	}
	for _, key := range []string{"$defs", "definitions"} {
		if d, ok := m[key].(map[string]any); ok {
			for name, c := range d {
				strictWalk(path+"/"+key+"/"+name, c, out)
			}
		}
	}
}
