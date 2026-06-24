package request

import (
	"encoding/json"
	"testing"

	"restdeck/internal/domain"
)

func TestImportOpenAPICollection(t *testing.T) {
	raw := `{
	  "openapi": "3.0.3",
	  "info": { "title": "Demo API", "version": "1.0.0" },
	  "servers": [{ "url": "{{baseUrl}}" }],
	  "paths": {
	    "/users": {
	      "get": {
	        "summary": "List users",
	        "parameters": [
	          { "name": "page", "in": "query", "schema": { "type": "integer", "example": 2 } },
	          { "name": "X-Trace", "in": "header", "schema": { "type": "string", "example": "abc" } }
	        ]
	      },
	      "post": {
	        "summary": "Create user",
	        "requestBody": {
	          "content": {
	            "application/json": {
	              "example": { "name": "Ada" }
	            }
	          }
	        }
	      }
	    }
	  }
	}`
	collection, err := ImportOpenAPI(raw)
	if err != nil {
		t.Fatal(err)
	}
	if collection.Name != "Demo API" {
		t.Fatalf("name = %q", collection.Name)
	}
	if len(collection.Requests) != 2 {
		t.Fatalf("requests = %d", len(collection.Requests))
	}
	get := collection.Requests[0]
	if get.Method != "GET" || get.URL != "{{baseUrl}}/users" {
		t.Fatalf("unexpected GET request: %#v", get)
	}
	if len(get.Params) != 1 || get.Params[0].Key != "page" || get.Params[0].Value != "2" {
		t.Fatalf("query params not imported: %#v", get.Params)
	}
	if len(get.Headers) != 1 || get.Headers[0].Key != "X-Trace" {
		t.Fatalf("headers not imported: %#v", get.Headers)
	}
	post := collection.Requests[1]
	if post.BodyMode != domain.BodyModeJSON || post.Body == "" {
		t.Fatalf("json body not imported: %#v", post)
	}
}

func TestExportOpenAPICollection(t *testing.T) {
	raw, err := ExportOpenAPI(domain.Collection{
		Name: "Exported API",
		Requests: []domain.Request{{
			Name:     "Create user",
			Method:   "POST",
			URL:      "{{baseUrl}}/users",
			BodyMode: domain.BodyModeJSON,
			Body:     `{"name":"Ada"}`,
			Params:   []domain.KeyValue{{Enabled: true, Key: "trace", Value: "1"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}
	paths := doc["paths"].(map[string]any)
	if _, ok := paths["/users"]; !ok {
		t.Fatalf("/users path missing: %s", raw)
	}
}
