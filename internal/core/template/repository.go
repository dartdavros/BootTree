package template

import (
	"context"
	"fmt"
	"io/fs"

	"boottree/internal/assets"
)

type EmbeddedRepository struct{ root fs.FS }

func NewEmbeddedRepository() EmbeddedRepository {
	root, err := fs.Sub(assets.FS, "templates")
	if err != nil {
		panic(err)
	}
	return EmbeddedRepository{root: root}
}

func (r EmbeddedRepository) Get(ctx context.Context, path string) (string, error) {
	_ = ctx
	data, err := fs.ReadFile(r.root, path)
	if err != nil {
		return "", fmt.Errorf("read template %q: %w", path, err)
	}
	return string(data), nil
}
