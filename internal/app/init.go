package app

import (
	"context"

	"boottree/internal/core/model"
	"boottree/internal/core/planner"
	"boottree/internal/core/scanner"
)

type InitPlanner struct {
	Scanner scanner.Service
	Planner planner.Service
}

func (p InitPlanner) BuildExecutionPlan(ctx context.Context, root string, preset model.Preset, options model.InitOptions) (model.TreeSnapshot, model.ExecutionPlan, error) {
	snapshot, err := p.Scanner.Scan(ctx, root)
	if err != nil {
		return model.TreeSnapshot{}, model.ExecutionPlan{}, err
	}

	plan, err := p.Planner.Build(snapshot, preset, options)
	if err != nil {
		return model.TreeSnapshot{}, model.ExecutionPlan{}, err
	}

	return snapshot, plan, nil
}
