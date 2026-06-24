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

func TestImportOpenAPIYAMLWithServerAndPathParams(t *testing.T) {
	raw := `
openapi: 3.0.3
info:
  title: YAML API
  version: 1.0.0
servers:
  - url: https://api.one.test
  - url: https://api.two.test
paths:
  /users/{id}:
    parameters:
      - name: id
        in: path
        schema:
          type: string
          example: u-1
    get:
      summary: Get user
      parameters:
        - name: include
          in: query
          schema:
            type: string
            enum: [profile, roles]
      responses:
        "200":
          description: OK
`
	info, err := InspectOpenAPI(raw)
	if err != nil {
		t.Fatalf("inspect yaml: %v", err)
	}
	if len(info.Servers) != 2 || info.Servers[1] != "https://api.two.test" {
		t.Fatalf("servers = %#v", info.Servers)
	}
	collection, err := ImportOpenAPIWithOptions(raw, domain.OpenAPIImportOptions{ServerURL: info.Servers[1]})
	if err != nil {
		t.Fatalf("import yaml: %v", err)
	}
	if len(collection.Requests) != 1 {
		t.Fatalf("requests = %d", len(collection.Requests))
	}
	req := collection.Requests[0]
	if req.URL != "https://api.two.test/users/{{id}}" {
		t.Fatalf("url = %q", req.URL)
	}
	if len(req.Params) != 2 {
		t.Fatalf("params = %#v", req.Params)
	}
}

func TestImportSwagger2Collection(t *testing.T) {
	raw := `{
	  "swagger": "2.0",
	  "info": { "title": "Swagger API", "version": "1.0.0" },
	  "host": "api.example.com",
	  "basePath": "/v1",
	  "schemes": ["https", "http"],
	  "paths": {
	    "/users/{id}": {
	      "parameters": [
	        { "name": "id", "in": "path", "type": "string", "default": "u-1" }
	      ],
	      "get": {
	        "summary": "Get user",
	        "parameters": [
	          { "name": "verbose", "in": "query", "type": "boolean", "default": true },
	          { "name": "X-Trace", "in": "header", "type": "string", "default": "trace-1" }
	        ]
	      },
	      "post": {
	        "summary": "Update user",
	        "parameters": [
	          {
	            "name": "payload",
	            "in": "body",
	            "schema": {
	              "type": "object",
	              "properties": {
	                "name": { "type": "string", "example": "Ada" }
	              }
	            }
	          }
	        ]
	      }
	    },
	    "/upload": {
	      "post": {
	        "summary": "Upload avatar",
	        "consumes": ["multipart/form-data"],
	        "parameters": [
	          { "name": "avatar", "in": "formData", "type": "file" },
	          { "name": "name", "in": "formData", "type": "string", "default": "Ada" }
	        ]
	      }
	    }
	  }
	}`
	info, err := InspectOpenAPI(raw)
	if err != nil {
		t.Fatalf("inspect swagger: %v", err)
	}
	if len(info.Servers) != 2 || info.Servers[0] != "https://api.example.com/v1" {
		t.Fatalf("servers = %#v", info.Servers)
	}
	collection, err := ImportOpenAPI(raw)
	if err != nil {
		t.Fatalf("import swagger: %v", err)
	}
	if collection.Name != "Swagger API" {
		t.Fatalf("name = %q", collection.Name)
	}
	if len(collection.Requests) != 3 {
		t.Fatalf("requests = %d", len(collection.Requests))
	}
	get := findRequestByURL(collection.Requests, "https://api.example.com/v1/users/{{id}}", "GET")
	if get.Method != "GET" || get.URL != "https://api.example.com/v1/users/{{id}}" {
		t.Fatalf("unexpected GET request: %#v", get)
	}
	if len(get.Params) != 2 || get.Params[0].Key != "id" || get.Params[1].Key != "verbose" {
		t.Fatalf("params = %#v", get.Params)
	}
	if len(get.Headers) != 1 || get.Headers[0].Key != "X-Trace" {
		t.Fatalf("headers = %#v", get.Headers)
	}
	post := findRequestByURL(collection.Requests, "https://api.example.com/v1/users/{{id}}", "POST")
	if post.BodyMode != domain.BodyModeJSON || post.Body == "" {
		t.Fatalf("body = %#v", post)
	}
	upload := findRequestByURL(collection.Requests, "https://api.example.com/v1/upload", "POST")
	if upload.BodyMode != domain.BodyModeForm {
		t.Fatalf("upload body mode = %q", upload.BodyMode)
	}
	if len(upload.FormItems) != 2 || upload.FormItems[0].Type != "file" || upload.FormItems[1].Value != "Ada" {
		t.Fatalf("form items = %#v", upload.FormItems)
	}
}

func findRequestByURL(requests []domain.Request, url, method string) domain.Request {
	for _, request := range requests {
		if request.URL == url && request.Method == method {
			return request
		}
	}
	return domain.Request{}
}
