package request

import (
	"context"
	"strings"
	"testing"
	"time"

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

func TestResolverSupportsTypedTimestampVariables(t *testing.T) {
	env := domain.Environment{Variables: []domain.KeyValue{{Enabled: true, Key: "ts", ValueType: "timestamp", TimestampFormat: "milliseconds"}}}
	resolver := NewResolver(env, nil)
	resolver.now = func() time.Time { return time.Unix(1700000000, 123000000).UTC() }

	got, err := resolver.ResolveWithError("{{ts}}")
	if err != nil {
		t.Fatalf("resolve timestamp: %v", err)
	}
	if got != "1700000000123" {
		t.Fatalf("timestamp = %q", got)
	}
}

func TestResolverReadsResponseJSONPathFromHistory(t *testing.T) {
	env := domain.Environment{Variables: []domain.KeyValue{{
		Enabled:         true,
		Key:             "userId",
		ValueType:       "responseJsonPath",
		SourceRequestID: "r1",
		JSONPath:        "$.items[0].id",
	}}}
	resolver := NewResolverWithOptions(env, nil, ResolverOptions{
		Context: context.Background(),
		HistoryLookup: func(ctx context.Context, requestID string) (domain.HistoryItem, bool, error) {
			return domain.HistoryItem{
				RequestID: requestID,
				Response:  domain.Response{Body: `{"items":[{"id":"u-1"}]}`},
				CreatedAt: time.Now(),
			}, true, nil
		},
	})

	got, err := resolver.ResolveWithError("id={{userId}}")
	if err != nil {
		t.Fatalf("resolve response variable: %v", err)
	}
	if got != "id=u-1" {
		t.Fatalf("resolved value = %q", got)
	}
}

func TestResolverRefreshesResponseVariableAfterTimeout(t *testing.T) {
	env := domain.Environment{Variables: []domain.KeyValue{{
		Enabled:             true,
		Key:                 "token",
		ValueType:           "responseJsonPath",
		SourceRequestID:     "login",
		JSONPath:            "$.token",
		ResponseStrategy:    "refreshAfter",
		RefreshAfterSeconds: 10,
	}}}
	refreshes := 0
	resolver := NewResolverWithOptions(env, nil, ResolverOptions{
		Context: context.Background(),
		HistoryLookup: func(ctx context.Context, requestID string) (domain.HistoryItem, bool, error) {
			return domain.HistoryItem{
				RequestID: requestID,
				Response:  domain.Response{Body: `{"token":"old"}`},
				CreatedAt: time.Unix(100, 0),
			}, true, nil
		},
		ResponseRefresh: func(ctx context.Context, requestID string, variables map[string]string) (domain.HistoryItem, error) {
			refreshes++
			return domain.HistoryItem{
				RequestID: requestID,
				Response:  domain.Response{Body: `{"token":"new"}`},
				CreatedAt: time.Unix(200, 0),
			}, nil
		},
	})
	resolver.now = func() time.Time { return time.Unix(120, 0) }

	got, err := resolver.ResolveWithError("{{token}}")
	if err != nil {
		t.Fatalf("resolve refreshed variable: %v", err)
	}
	if got != "new" || refreshes != 1 {
		t.Fatalf("got %q refreshes %d", got, refreshes)
	}
}

func TestResolverDetectsVariableCycles(t *testing.T) {
	env := domain.Environment{Variables: []domain.KeyValue{
		{Enabled: true, Key: "a", Value: "{{b}}"},
		{Enabled: true, Key: "b", Value: "{{a}}"},
	}}
	_, err := NewResolver(env, nil).ResolveWithError("{{a}}")
	if err == nil {
		t.Fatal("expected cycle error")
	}
}
