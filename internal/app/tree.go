package app

import (
	"context"

	"boottree/internal/core/model"
	"boottree/internal/core/scanner"
)

type TreeOptions struct {
	IncludeIgnored bool
}

type TreeBuilder struct {
	Scanner scanner.Service
}

func (b TreeBuilder) BuildSnapshot(ctx context.Context, root string, options TreeOptions) (model.TreeSnapshot, error) {
	return b.Scanner.ScanWithOptions(ctx, root, scanner.Options{IncludeIgnored: options.IncludeIgnored})
}
