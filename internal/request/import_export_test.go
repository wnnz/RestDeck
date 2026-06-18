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
