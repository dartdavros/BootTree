package template

import (
	"context"
	"fmt"
	"io/fs"

	templatesfs "boottree/templates"
)

type EmbeddedRepository struct{ root fs.FS }

func NewEmbeddedRepository() EmbeddedRepository {
	root, err := fs.Sub(templatesfs.FS, ".")
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
