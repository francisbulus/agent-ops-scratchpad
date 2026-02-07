package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

// Error represents one schema validation failure.
type Error struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// EventValidator validates telemetry payloads against the v0 event schema.
type EventValidator struct {
	schema map[string]any
}

// NewEventValidator loads and parses a JSON schema file.
func NewEventValidator(schemaPath string) (*EventValidator, error) {
	resolvedPath, err := resolveSchemaPath(schemaPath)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("parse schema json: %w", err)
	}

	return &EventValidator{schema: schema}, nil
}

// Validate returns all schema violations for a decoded JSON payload.
func (v *EventValidator) Validate(payload any) []Error {
	if v == nil {
		return []Error{{Path: "$", Message: "validator is not configured"}}
	}

	errList := make([]Error, 0)
	validateNode(v.schema, payload, "$", &errList)
	return errList
}

func resolveSchemaPath(schemaPath string) (string, error) {
	if strings.TrimSpace(schemaPath) == "" {
		schemaPath = "packages/schemas/agent-event-v0.schema.json"
	}

	candidates := []string{schemaPath}
	if !filepath.IsAbs(schemaPath) {
		candidates = append(candidates,
			filepath.Join("..", schemaPath),
			filepath.Join("..", "..", schemaPath),
			filepath.Join("..", "..", "..", schemaPath),
		)
	}

	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			abs, err := filepath.Abs(candidate)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("schema file not found from path %q", schemaPath)
}

func validateNode(schema map[string]any, value any, path string, errs *[]Error) {
	if schema == nil {
		return
	}

	if !matchesType(schema["type"], value) {
		if typeSpec, ok := schema["type"]; ok {
			addErr(errs, path, fmt.Sprintf("must be type %s", typeDescription(typeSpec)))
			return
		}
	}

	if enumVals, ok := toSlice(schema["enum"]); ok && !valueInEnum(value, enumVals) {
		addErr(errs, path, "must be one of allowed enum values")
	}

	if constVal, ok := schema["const"]; ok && !equalJSONValue(value, constVal) {
		addErr(errs, path, fmt.Sprintf("must equal %v", constVal))
	}

	if minRaw, ok := toFloat(schema["minimum"]); ok {
		num, numOK := toFloat(value)
		if !numOK || num < minRaw {
			addErr(errs, path, fmt.Sprintf("must be >= %s", trimFloat(minRaw)))
		}
	}

	if minLength, ok := toInt(schema["minLength"]); ok {
		str, ok := value.(string)
		if !ok || len(str) < minLength {
			addErr(errs, path, fmt.Sprintf("must have length >= %d", minLength))
		}
	}

	if format, ok := schema["format"].(string); ok {
		validateFormat(format, value, path, errs)
	}

	obj, isObj := value.(map[string]any)
	if isObj {
		if req, ok := toStringSlice(schema["required"]); ok {
			for _, key := range req {
				if _, found := obj[key]; !found {
					addErr(errs, childPath(path, key), "is required")
				}
			}
		}

		if props, ok := toMap(schema["properties"]); ok {
			for key, val := range obj {
				if propSchema, found := toMap(props[key]); found {
					validateNode(propSchema, val, childPath(path, key), errs)
					continue
				}

				handleAdditionalProperties(schema["additionalProperties"], val, childPath(path, key), errs)
			}
		}
	}

	if allOfItems, ok := toSlice(schema["allOf"]); ok {
		for _, item := range allOfItems {
			itemMap, ok := toMap(item)
			if !ok {
				continue
			}

			ifSchema, hasIf := toMap(itemMap["if"])
			thenSchema, hasThen := toMap(itemMap["then"])
			if !hasThen {
				continue
			}

			if !hasIf || matches(ifSchema, value) {
				validateNode(thenSchema, value, path, errs)
			}
		}
	}
}

func matches(schema map[string]any, value any) bool {
	errList := make([]Error, 0)
	validateNode(schema, value, "$", &errList)
	return len(errList) == 0
}

func validateFormat(format string, value any, path string, errs *[]Error) {
	str, ok := value.(string)
	if !ok {
		addErr(errs, path, fmt.Sprintf("must match %s format", format))
		return
	}

	switch format {
	case "date-time":
		if _, err := time.Parse(time.RFC3339, str); err != nil {
			addErr(errs, path, "must be RFC3339 date-time")
		}
	case "uuid":
		if !uuidPattern.MatchString(str) {
			addErr(errs, path, "must be a valid UUID")
		}
	}
}

func handleAdditionalProperties(additional any, value any, path string, errs *[]Error) {
	if additional == nil {
		return
	}

	if allowed, ok := additional.(bool); ok {
		if !allowed {
			addErr(errs, path, "additional property is not allowed")
		}
		return
	}

	if schema, ok := toMap(additional); ok {
		validateNode(schema, value, path, errs)
	}
}

func matchesType(typeSpec any, value any) bool {
	if typeSpec == nil {
		return true
	}

	switch t := typeSpec.(type) {
	case string:
		return matchesSingleType(t, value)
	case []any:
		for _, raw := range t {
			name, ok := raw.(string)
			if ok && matchesSingleType(name, value) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func matchesSingleType(typeName string, value any) bool {
	switch typeName {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "integer":
		return isInteger(value)
	case "number":
		_, ok := toFloat(value)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	case "array":
		_, ok := value.([]any)
		return ok
	default:
		return true
	}
}

func isInteger(value any) bool {
	switch n := value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float64:
		return n == float64(int64(n))
	case float32:
		return float64(n) == float64(int64(n))
	case json.Number:
		_, err := n.Int64()
		return err == nil
	default:
		return false
	}
}

func valueInEnum(value any, enumValues []any) bool {
	for _, allowed := range enumValues {
		if equalJSONValue(value, allowed) {
			return true
		}
	}
	return false
}

func equalJSONValue(a, b any) bool {
	if fa, ok := toFloat(a); ok {
		if fb, ok := toFloat(b); ok {
			return fa == fb
		}
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func toMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

func toSlice(v any) ([]any, bool) {
	s, ok := v.([]any)
	return s, ok
}

func toStringSlice(v any) ([]string, bool) {
	raw, ok := toSlice(v)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		str, ok := item.(string)
		if !ok {
			continue
		}
		out = append(out, str)
	}
	return out, true
}

func toInt(v any) (int, bool) {
	n, ok := toFloat(v)
	if !ok {
		return 0, false
	}
	return int(n), true
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func addErr(errs *[]Error, path string, message string) {
	*errs = append(*errs, Error{Path: path, Message: message})
}

func childPath(parent string, field string) string {
	if parent == "$" {
		return "$." + field
	}
	return parent + "." + field
}

func trimFloat(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func typeDescription(typeSpec any) string {
	switch t := typeSpec.(type) {
	case string:
		return t
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		return strings.Join(parts, "|")
	default:
		return "unknown"
	}
}
