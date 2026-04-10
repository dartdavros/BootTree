package templates

import (
    "io/fs"
    "testing"
)

func TestFS_ContainsSoftwareProductTemplates(t *testing.T) {
    t.Parallel()

    required := []string{
        "software-product/docs/README.md.tmpl",
        "software-product/engineering/architecture.md.tmpl",
    }

    for _, path := range required {
        data, err := fs.ReadFile(FS, path)
        if err != nil {
            t.Fatalf("read embedded template %q: %v", path, err)
        }
        if len(data) == 0 {
            t.Fatalf("embedded template %q is empty", path)
        }
    }
}
