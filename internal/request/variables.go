package request

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"restdeck/internal/domain"
)

var variablePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

type HistoryLookup func(context.Context, string) (domain.HistoryItem, bool, error)
type ResponseRefresh func(context.Context, string, map[string]string) (domain.HistoryItem, error)

type ResolverOptions struct {
	Context         context.Context
	HistoryLookup   HistoryLookup
	ResponseRefresh ResponseRefresh
}

type Resolver struct {
	values    map[string]string
	variables map[string]domain.KeyValue
	now       func() time.Time
	options   ResolverOptions
	resolving map[string]bool
}

func NewResolver(env domain.Environment, globals []domain.KeyValue) *Resolver {
	return NewResolverWithOptions(env, globals, ResolverOptions{})
}

func NewResolverWithOptions(env domain.Environment, globals []domain.KeyValue, options ResolverOptions) *Resolver {
	values := map[string]string{}
	variables := map[string]domain.KeyValue{}
	for _, kv := range globals {
		if kv.Enabled && kv.Key != "" {
			kv = normalizeVariable(kv)
			variables[kv.Key] = kv
			if kv.ValueType == "" || kv.ValueType == "static" {
				values[kv.Key] = kv.Value
			}
		}
	}
	for _, kv := range env.Variables {
		if kv.Enabled && kv.Key != "" {
			kv = normalizeVariable(kv)
			variables[kv.Key] = kv
			if kv.ValueType == "" || kv.ValueType == "static" {
				values[kv.Key] = kv.Value
			}
		}
	}
	if options.Context == nil {
		options.Context = context.Background()
	}
	return &Resolver{values: values, variables: variables, now: time.Now, options: options, resolving: map[string]bool{}}
}

func (r *Resolver) Resolve(input string) string {
	value, err := r.ResolveWithError(input)
	if err != nil {
		return input
	}
	return value
}

func (r *Resolver) ResolveWithError(input string) (string, error) {
	if input == "" {
		return input, nil
	}
	var resolveErr error
	out := variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := variablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		name := strings.TrimSpace(parts[1])
		if value, ok := r.dynamic(name); ok {
			return value
		}
		value, ok, err := r.resolveVariable(name)
		if err != nil {
			resolveErr = err
			return match
		}
		if ok {
			return value
		}
		return match
	})
	if resolveErr != nil {
		return input, resolveErr
	}
	return out, nil
}

func (r *Resolver) Values() map[string]string {
	out, err := r.ValuesWithError()
	if err != nil {
		return map[string]string{}
	}
	return out
}

func (r *Resolver) ValuesWithError() (map[string]string, error) {
	out := map[string]string{}
	for key := range r.variables {
		value, ok, err := r.resolveVariable(key)
		if err != nil {
			return nil, err
		}
		if ok {
			out[key] = value
		}
	}
	for key, value := range r.values {
		if _, ok := out[key]; !ok {
			out[key] = value
		}
	}
	return out, nil
}

func (r *Resolver) resolveVariable(name string) (string, bool, error) {
	if value, ok := r.values[name]; ok {
		if variablePattern.MatchString(value) {
			resolved, err := r.resolveNamedValue(name, value)
			return resolved, true, err
		}
		return value, true, nil
	}
	kv, ok := r.variables[name]
	if !ok {
		return "", false, nil
	}
	value, err := r.resolveKeyValue(kv)
	if err != nil {
		return "", false, err
	}
	r.values[name] = value
	return value, true, nil
}

func (r *Resolver) resolveNamedValue(name, value string) (string, error) {
	if r.resolving[name] {
		return "", fmt.Errorf("variable cycle detected at %q", name)
	}
	r.resolving[name] = true
	defer delete(r.resolving, name)
	return r.ResolveWithError(value)
}

func (r *Resolver) resolveKeyValue(kv domain.KeyValue) (string, error) {
	if r.resolving[kv.Key] {
		return "", fmt.Errorf("variable cycle detected at %q", kv.Key)
	}
	r.resolving[kv.Key] = true
	defer delete(r.resolving, kv.Key)
	switch kv.ValueType {
	case "", "static":
		return r.ResolveWithError(kv.Value)
	case "timestamp":
		return r.timestamp(kv.TimestampFormat), nil
	case "responseJsonPath":
		value, ok, err := r.responseJSONPath(kv)
		if err != nil {
			return "", err
		}
		if ok {
			return value, nil
		}
		if kv.FallbackValue != "" {
			return r.ResolveWithError(kv.FallbackValue)
		}
		return "{{" + kv.Key + "}}", nil
	default:
		return r.ResolveWithError(kv.Value)
	}
}

func (r *Resolver) dynamic(name string) (string, bool) {
	now := r.now().UTC()
	switch name {
	case "$guid", "$randomUUID":
		return uuid.NewString(), true
	case "$timestamp":
		return strconv.FormatInt(now.Unix(), 10), true
	case "$isoTimestamp":
		return now.Format(time.RFC3339), true
	case "$randomInt":
		return strconv.Itoa(mathrand.Intn(1000)), true
	case "$randomBoolean":
		return strconv.FormatBool(mathrand.Intn(2) == 0), true
	case "$randomEmail":
		return fmt.Sprintf("user_%s@example.com", randomHex(4)), true
	case "$randomUserName":
		return "user_" + randomHex(4), true
	case "$randomFirstName":
		names := []string{"Alex", "Ming", "Sam", "River", "Nora"}
		return names[mathrand.Intn(len(names))], true
	default:
		return "", false
	}
}

func (r *Resolver) timestamp(format string) string {
	now := r.now().UTC()
	switch format {
	case "milliseconds":
		return strconv.FormatInt(now.UnixMilli(), 10)
	case "iso":
		return now.Format(time.RFC3339)
	default:
		return strconv.FormatInt(now.Unix(), 10)
	}
}

func (r *Resolver) responseJSONPath(kv domain.KeyValue) (string, bool, error) {
	if kv.SourceRequestID == "" || kv.JSONPath == "" || r.options.HistoryLookup == nil {
		return "", false, nil
	}
	ctx := r.options.Context
	history, found, err := r.options.HistoryLookup(ctx, kv.SourceRequestID)
	if err != nil {
		return "", false, err
	}
	needsRefresh := kv.ResponseStrategy == "alwaysRequest" || (kv.ResponseStrategy == "refreshAfter" && !found)
	if kv.ResponseStrategy == "refreshAfter" && found && kv.RefreshAfterSeconds > 0 {
		needsRefresh = r.now().Sub(history.CreatedAt) > time.Duration(kv.RefreshAfterSeconds)*time.Second
	}
	if needsRefresh && r.options.ResponseRefresh != nil {
		variables, err := r.valuesExcluding(kv.Key)
		if err != nil {
			return "", false, err
		}
		history, err = r.options.ResponseRefresh(ctx, kv.SourceRequestID, variables)
		if err != nil {
			return "", false, err
		}
		found = true
	}
	if !found {
		return "", false, nil
	}
	return extractJSONPath(history.Response.Body, kv.JSONPath)
}

func (r *Resolver) valuesExcluding(excluded string) (map[string]string, error) {
	out := map[string]string{}
	for key := range r.variables {
		if key == excluded {
			continue
		}
		value, ok, err := r.resolveVariable(key)
		if err != nil {
			return nil, err
		}
		if ok {
			out[key] = value
		}
	}
	for key, value := range r.values {
		if key == excluded {
			continue
		}
		if _, ok := out[key]; !ok {
			out[key] = value
		}
	}
	return out, nil
}

func extractJSONPath(raw, path string) (string, bool, error) {
	var value interface{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return "", false, err
	}
	current := value
	rest := strings.TrimSpace(path)
	if rest == "$" {
		return stringifyJSONPathValue(current), true, nil
	}
	if !strings.HasPrefix(rest, "$.") && !strings.HasPrefix(rest, "$[") {
		return "", false, fmt.Errorf("unsupported JSONPath %q", path)
	}
	rest = strings.TrimPrefix(rest, "$")
	for rest != "" {
		if strings.HasPrefix(rest, ".") {
			rest = rest[1:]
			end := strings.IndexAny(rest, ".[")
			key := rest
			if end >= 0 {
				key = rest[:end]
				rest = rest[end:]
			} else {
				rest = ""
			}
			obj, ok := current.(map[string]interface{})
			if !ok {
				return "", false, nil
			}
			next, ok := obj[key]
			if !ok {
				return "", false, nil
			}
			current = next
			continue
		}
		if strings.HasPrefix(rest, "[") {
			end := strings.Index(rest, "]")
			if end < 0 {
				return "", false, fmt.Errorf("unsupported JSONPath %q", path)
			}
			index, err := strconv.Atoi(rest[1:end])
			if err != nil {
				return "", false, fmt.Errorf("unsupported JSONPath %q", path)
			}
			list, ok := current.([]interface{})
			if !ok || index < 0 || index >= len(list) {
				return "", false, nil
			}
			current = list[index]
			rest = rest[end+1:]
			continue
		}
		return "", false, fmt.Errorf("unsupported JSONPath %q", path)
	}
	return stringifyJSONPathValue(current), true, nil
}

func stringifyJSONPathValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case nil:
		return "null"
	case float64, bool:
		return fmt.Sprint(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(data)
	}
}

func normalizeVariable(kv domain.KeyValue) domain.KeyValue {
	if kv.ValueType == "" {
		kv.ValueType = "static"
	}
	if kv.ResponseStrategy == "" {
		kv.ResponseStrategy = "latestHistory"
	}
	return kv
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return strconv.Itoa(mathrand.Intn(100000))
	}
	return hex.EncodeToString(buf)
}
