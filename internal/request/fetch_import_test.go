package request

import (
	"strings"
	"testing"

	"restdeck/internal/domain"
)

func TestImportFetchGET(t *testing.T) {
	req, err := ImportFetch(`fetch("https://api.example.com/v1/users?page=1", {
	  "headers": {
	    "accept": "application/json",
	    "x-trace": "abc"
	  },
	  "credentials": "include"
	});`, "c1", 3)
	if err != nil {
		t.Fatalf("import fetch: %v", err)
	}

	if req.CollectionID != "c1" {
		t.Fatalf("collection id = %q", req.CollectionID)
	}
	if req.Method != "GET" {
		t.Fatalf("method = %q", req.Method)
	}
	if req.URL != "https://api.example.com/v1/users?page=1" {
		t.Fatalf("url = %q", req.URL)
	}
	if req.Name != "GET api.example.com/v1/users" {
		t.Fatalf("name = %q", req.Name)
	}
	if req.BodyMode != domain.BodyModeNone {
		t.Fatalf("body mode = %q", req.BodyMode)
	}
	if len(req.Headers) != 2 || req.Headers[0].Key != "accept" || req.Headers[0].Value != "application/json" {
		t.Fatalf("headers = %#v", req.Headers)
	}
	if req.SortOrder != 3 {
		t.Fatalf("sort order = %d", req.SortOrder)
	}
}

func TestImportFetchPOSTJSON(t *testing.T) {
	req, err := ImportFetch(`await fetch('https://api.example.com/v1/users', {
	  method: 'POST',
	  headers: new Headers({
	    'Content-Type': 'application/json',
	    Authorization: 'Bearer token'
	  }),
	  body: "{\"name\":\"Ada\"}",
	  credentials: "include"
	})`, "c1", 0)
	if err != nil {
		t.Fatalf("import fetch: %v", err)
	}

	if req.Method != "POST" {
		t.Fatalf("method = %q", req.Method)
	}
	if req.BodyMode != domain.BodyModeJSON {
		t.Fatalf("body mode = %q", req.BodyMode)
	}
	if req.Body != `{"name":"Ada"}` {
		t.Fatalf("body = %q", req.Body)
	}
	if req.Auth.Type != domain.AuthTypeNone {
		t.Fatalf("auth = %#v", req.Auth)
	}
	if len(req.Headers) != 2 {
		t.Fatalf("headers = %#v", req.Headers)
	}
	if req.Headers[1].Key != "Authorization" || req.Headers[1].Value != "Bearer token" {
		t.Fatalf("authorization header was not preserved: %#v", req.Headers)
	}
}

func TestImportFetchURLEncodedBody(t *testing.T) {
	req, err := ImportFetch(`fetch("https://api.example.com/session", {
	  "method": "POST",
	  "headers": {"content-type": "application/x-www-form-urlencoded;charset=UTF-8"},
	  "body": "user=ada&role=admin"
	})`, "c1", 0)
	if err != nil {
		t.Fatalf("import fetch: %v", err)
	}

	if req.BodyMode != domain.BodyModeURLEncoded {
		t.Fatalf("body mode = %q", req.BodyMode)
	}
	if req.Body != "user=ada\nrole=admin" {
		t.Fatalf("body = %q", req.Body)
	}
}

func TestImportFetchRejectsNonFetchText(t *testing.T) {
	_, err := ImportFetch(`console.log("not fetch")`, "c1", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "fetch") {
		t.Fatalf("error = %v", err)
	}
}
