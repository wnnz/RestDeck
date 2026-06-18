package request

import (
	"strings"
	"testing"

	"restdeck/internal/domain"
)

func TestResolverUsesEnvironmentAndDynamicVariables(t *testing.T) {
	env := domain.Environment{Variables: []domain.KeyValue{{Enabled: true, Key: "baseUrl", Value: "https://api.example.com"}}}
	resolver := NewResolver(env, []domain.KeyValue{{Enabled: true, Key: "team", Value: "core"}})

	got := resolver.Resolve("{{baseUrl}}/{{team}}/{{$guid}}")

	if !strings.HasPrefix(got, "https://api.example.com/core/") {
		t.Fatalf("unexpected resolved value: %s", got)
	}
	if strings.Contains(got, "{{") {
		t.Fatalf("dynamic variable was not resolved: %s", got)
	}
}
