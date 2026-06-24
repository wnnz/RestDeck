package request

import (
	"strings"
	"testing"
)

func TestPostmanImportExportRoundTrip(t *testing.T) {
	raw := `{
	  "info": {"name": "Demo"},
	  "item": [{
	    "name": "Get user",
	    "request": {
	      "method": "GET",
	      "url": "https://api.example.com/users/{{id}}",
	      "header": [{"key": "Accept", "value": "application/json"}],
	      "auth": {"type": "bearer", "bearer": [{"key": "token", "value": "{{token}}"}]},
	      "event": [{"listen": "test", "script": {"type": "text/javascript", "exec": ["pm.test(\"ok\", function () { expect(pm.response.code).to.equal(200); });"]}}]
	    }
	  }]
	}`

	collection, err := ImportPostman(raw)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if collection.Name != "Demo" || len(collection.Requests) != 1 {
		t.Fatalf("unexpected import result: %#v", collection)
	}
	if collection.Requests[0].Auth.Type != "bearer" {
		t.Fatalf("auth was not imported")
	}

	exported, err := ExportPostman(collection)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if !strings.Contains(exported, "Get user") || !strings.Contains(exported, "bearer") {
		t.Fatalf("export missing expected content: %s", exported)
	}
}

func TestPostmanImportURLQueryAndFolderInheritance(t *testing.T) {
	raw := `{
	  "info": {"name": "Demo"},
	  "item": [{
	    "name": "Folder",
	    "auth": {"type": "bearer", "bearer": [{"key": "token", "value": "{{folderToken}}"}]},
	    "event": [{"listen": "prerequest", "script": {"type": "text/javascript", "exec": ["pm.variables.set(\"folder\", \"1\");"]}}],
	    "item": [{
	      "name": "Get user",
	      "request": {
	        "method": "GET",
	        "url": {
	          "protocol": "https",
	          "host": ["api", "example", "com"],
	          "path": ["users"],
	          "query": [{"key": "page", "value": "1"}]
	        }
	      }
	    }]
	  }]
	}`

	collection, err := ImportPostman(raw)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if len(collection.Requests) != 1 {
		t.Fatalf("requests = %d", len(collection.Requests))
	}
	req := collection.Requests[0]
	if len(req.Params) != 1 || req.Params[0].Key != "page" || req.Params[0].Value != "1" {
		t.Fatalf("query params = %#v", req.Params)
	}
	if req.Auth.Type != "bearer" || req.Auth.Values["token"] != "{{folderToken}}" {
		t.Fatalf("auth = %#v", req.Auth)
	}
	if !strings.Contains(req.PreScript, "folder") {
		t.Fatalf("pre script = %q", req.PreScript)
	}
}
