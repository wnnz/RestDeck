package request

import (
	"crypto/rand"
	"encoding/hex"
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

type Resolver struct {
	values map[string]string
	now    func() time.Time
}

func NewResolver(env domain.Environment, globals []domain.KeyValue) *Resolver {
	values := map[string]string{}
	for _, kv := range globals {
		if kv.Enabled && kv.Key != "" {
			values[kv.Key] = kv.Value
		}
	}
	for _, kv := range env.Variables {
		if kv.Enabled && kv.Key != "" {
			values[kv.Key] = kv.Value
		}
	}
	return &Resolver{values: values, now: time.Now}
}

func (r *Resolver) Resolve(input string) string {
	if input == "" {
		return input
	}
	return variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := variablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		name := strings.TrimSpace(parts[1])
		if value, ok := r.dynamic(name); ok {
			return value
		}
		if value, ok := r.values[name]; ok {
			return value
		}
		return match
	})
}

func (r *Resolver) Values() map[string]string {
	out := map[string]string{}
	for k, v := range r.values {
		out[k] = v
	}
	return out
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

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return strconv.Itoa(mathrand.Intn(100000))
	}
	return hex.EncodeToString(buf)
}
