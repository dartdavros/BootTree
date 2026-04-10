package app

import (
	"context"

	"boottree/internal/core/model"
	corestats "boottree/internal/core/stats"
	"boottree/internal/core/scanner"
)

type StatsBuilder struct {
	Scanner scanner.Service
	Stats   corestats.Service
}

func (b StatsBuilder) Build(ctx context.Context, root string) (model.ProjectStats, error) {
	snapshot, err := b.Scanner.Scan(ctx, root)
	if err != nil {
		return model.ProjectStats{}, err
	}
	return b.Stats.Build(snapshot), nil
}
