package request

import (
	"testing"

	"restdeck/internal/domain"
)

func TestScriptRuntimeRunsPreRequestAndTests(t *testing.T) {
	runtime := NewScriptRuntime()
	vars := map[string]string{"base": "ready"}
	pre := runtime.RunPreRequest(t.Context(), `pm.variables.set("trace", "abc");`, domain.Request{}, vars)
	if len(pre) != 0 {
		t.Fatalf("unexpected pre-request results: %#v", pre)
	}
	if vars["trace"] != "abc" {
		t.Fatalf("pre-request did not mutate variables")
	}

	results := runtime.RunTests(t.Context(),
		`pm.test("status", function () { expect(pm.response.code).to.equal(201); });`,
		domain.Request{},
		domain.Response{StatusCode: 201, Body: `{"ok":true}`},
		vars,
	)
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("unexpected test results: %#v", results)
	}
}
