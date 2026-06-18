package request

import (
	"context"
	"time"

	"github.com/google/uuid"

	"restdeck/internal/domain"
)

type Runner struct {
	sender  *Sender
	scripts *ScriptRuntime
}

func NewRunner(sender *Sender, scripts *ScriptRuntime) *Runner {
	return &Runner{sender: sender, scripts: scripts}
}

func (r *Runner) Run(ctx context.Context, collection domain.Collection, env domain.Environment, globals []domain.KeyValue, iterations int) domain.RunnerResult {
	if iterations <= 0 {
		iterations = 1
	}
	start := time.Now()
	result := domain.RunnerResult{
		ID:            uuid.NewString(),
		CollectionID:  collection.ID,
		EnvironmentID: env.ID,
		Name:          collection.Name,
		Iterations:    iterations,
		CreatedAt:     time.Now(),
	}
	variables := NewResolver(env, globals).Values()
	for i := 0; i < iterations; i++ {
		for _, req := range collection.Requests {
			preResults := r.scripts.RunPreRequest(ctx, req.PreScript, req, variables)
			res, err := r.sender.SendWithVariables(ctx, req, variables)
			if err != nil {
				item := domain.TestResult{Name: req.Name, Passed: false, Message: err.Error()}
				result.Items = append(result.Items, item)
				result.Failed++
				continue
			}
			tests := r.scripts.RunTests(ctx, req.TestScript, req, res, variables)
			if len(preResults) > 0 {
				tests = append(preResults, tests...)
			}
			if len(tests) == 0 {
				item := domain.TestResult{Name: req.Name, Passed: res.StatusCode >= 200 && res.StatusCode < 400, Message: res.Status}
				result.Items = append(result.Items, item)
				if item.Passed {
					result.Passed++
				} else {
					result.Failed++
				}
				continue
			}
			for _, test := range tests {
				test.Name = req.Name + " / " + test.Name
				result.Items = append(result.Items, test)
				if test.Passed {
					result.Passed++
				} else {
					result.Failed++
				}
			}
		}
	}
	result.DurationMs = time.Since(start).Milliseconds()
	return result
}
