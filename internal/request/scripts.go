package request

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"

	"restdeck/internal/domain"
)

type ScriptRuntime struct{}

func NewScriptRuntime() *ScriptRuntime {
	return &ScriptRuntime{}
}

func (r *ScriptRuntime) RunPreRequest(ctx context.Context, script string, req domain.Request, variables map[string]string) []domain.TestResult {
	if script == "" {
		return nil
	}
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	_ = vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value { return goja.Undefined() },
	})
	pm := map[string]interface{}{
		"request": req,
		"variables": map[string]interface{}{
			"get": func(name string) string { return variables[name] },
			"set": func(name, value string) { variables[name] = value },
			"replaceIn": func(value string) string {
				return resolveFromMap(value, variables)
			},
		},
	}
	_ = vm.Set("pm", pm)
	err := runWithTimeout(ctx, vm, script)
	if err != nil {
		return []domain.TestResult{{Name: "Pre-request error", Passed: false, Message: err.Error()}}
	}
	return nil
}

func (r *ScriptRuntime) RunTests(ctx context.Context, script string, req domain.Request, res domain.Response, variables map[string]string) []domain.TestResult {
	if script == "" {
		return nil
	}
	results := []domain.TestResult{}
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	console := map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value { return goja.Undefined() },
	}
	_ = vm.Set("console", console)
	pm := map[string]interface{}{}
	pm["variables"] = map[string]interface{}{
		"get": func(name string) string { return variables[name] },
		"set": func(name, value string) { variables[name] = value },
		"replaceIn": func(value string) string {
			return resolveFromMap(value, variables)
		},
	}
	pm["request"] = req
	pm["response"] = map[string]interface{}{
		"code":         res.StatusCode,
		"status":       res.Status,
		"headers":      headersMap(res.Headers),
		"responseTime": res.DurationMs,
		"text": func() string {
			return res.Body
		},
		"json": func() interface{} {
			var out interface{}
			if err := gojaJSON(vm, res.Body, &out); err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return out
		},
	}
	pm["test"] = func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		passed := true
		message := ""
		if len(call.Arguments) > 1 {
			if callable, ok := goja.AssertFunction(call.Arguments[1]); ok {
				_, err := callable(goja.Undefined())
				if err != nil {
					passed = false
					message = err.Error()
				}
			}
		}
		results = append(results, domain.TestResult{Name: name, Passed: passed, Message: message})
		return goja.Undefined()
	}
	_ = vm.Set("pm", pm)
	_ = vm.Set("expect", expectFactory(vm))

	if err := runWithTimeout(ctx, vm, script); err != nil {
		results = append(results, domain.TestResult{Name: "Script error", Passed: false, Message: err.Error()})
	}
	return results
}

func runWithTimeout(ctx context.Context, vm *goja.Runtime, script string) error {
	timeout := time.AfterFunc(5*time.Second, func() {
		vm.Interrupt("script exceeded 5 seconds")
	})
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt(ctx.Err().Error())
		case <-done:
		}
	}()
	_, err := vm.RunString(script)
	close(done)
	timeout.Stop()
	return err
}

func resolveFromMap(value string, variables map[string]string) string {
	fakeEnv := domain.Environment{Variables: []domain.KeyValue{}}
	for key, val := range variables {
		fakeEnv.Variables = append(fakeEnv.Variables, domain.KeyValue{Enabled: true, Key: key, Value: val})
	}
	return NewResolver(fakeEnv, nil).Resolve(value)
}

func headersMap(headers []domain.KeyValue) map[string]string {
	out := map[string]string{}
	for _, header := range headers {
		out[header.Key] = header.Value
	}
	return out
}

func gojaJSON(vm *goja.Runtime, raw string, out *interface{}) error {
	value := vm.Get("JSON").ToObject(vm).Get("parse")
	parse, ok := goja.AssertFunction(value)
	if !ok {
		return fmt.Errorf("JSON.parse is unavailable")
	}
	parsed, err := parse(goja.Undefined(), vm.ToValue(raw))
	if err != nil {
		return err
	}
	*out = parsed.Export()
	return nil
}

func expectFactory(vm *goja.Runtime) func(goja.Value) map[string]interface{} {
	return func(actual goja.Value) map[string]interface{} {
		assert := func(condition bool, message string) {
			if !condition {
				panic(vm.ToValue(message))
			}
		}
		return map[string]interface{}{
			"to": map[string]interface{}{
				"equal": func(expected goja.Value) {
					assert(actual.StrictEquals(expected), fmt.Sprintf("expected %v to equal %v", actual, expected))
				},
				"include": func(expected string) {
					assert(stringsContains(actual.String(), expected), fmt.Sprintf("expected %q to include %q", actual.String(), expected))
				},
				"be": map[string]interface{}{
					"ok": func() {
						assert(actual.ToBoolean(), "expected value to be truthy")
					},
				},
			},
		}
	}
}

func stringsContains(value, expected string) bool {
	return strings.Contains(value, expected)
}
