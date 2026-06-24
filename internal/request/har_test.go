package request

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestImportHARCollection(t *testing.T) {
	raw := `{
	  "log": {
	    "version": "1.2",
	    "creator": { "name": "browser", "version": "1" },
	    "entries": [{
	      "request": {
	        "method": "POST",
	        "url": "https://api.example.com/users",
	        "headers": [{"name": "Content-Type", "value": "application/json"}],
	        "queryString": [{"name": "trace", "value": "1"}],
	        "postData": { "mimeType": "application/json", "text": "{\"name\":\"Ada\"}" }
	      },
	      "response": { "status": 201, "statusText": "Created", "headers": [], "content": {"size": 0, "mimeType": ""} }
	    }]
	  }
	}`
	collection, err := ImportHAR(raw)
	if err != nil {
		t.Fatalf("import har: %v", err)
	}
	if len(collection.Requests) != 1 {
		t.Fatalf("requests = %d", len(collection.Requests))
	}
	req := collection.Requests[0]
	if req.Method != "POST" || req.BodyMode != "json" || !strings.Contains(req.Body, "Ada") {
		t.Fatalf("request = %#v", req)
	}
}

func TestExportHARCollection(t *testing.T) {
	collection, err := ImportHAR(`{
	  "log": {
	    "entries": [{
	      "request": {
	        "method": "GET",
	        "url": "https://api.example.com/users",
	        "headers": [],
	        "queryString": []
	      },
	      "response": { "status": 200, "statusText": "OK", "headers": [], "content": {"size": 0, "mimeType": ""} }
	    }]
	  }
	}`)
	if err != nil {
		t.Fatalf("import seed har: %v", err)
	}
	raw, err := ExportHAR(collection)
	if err != nil {
		t.Fatalf("export har: %v", err)
	}
	var archive map[string]any
	if err := json.Unmarshal([]byte(raw), &archive); err != nil {
		t.Fatalf("exported har json: %v", err)
	}
	if !strings.Contains(raw, "https://api.example.com/users") {
		t.Fatalf("exported har missing url: %s", raw)
	}
}
