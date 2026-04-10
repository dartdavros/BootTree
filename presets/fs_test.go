package presets

import (
    "io/fs"
    "testing"
)

func TestFS_ContainsSoftwareProductPreset(t *testing.T) {
    t.Parallel()

    data, err := fs.ReadFile(FS, "software-product/preset.json")
    if err != nil {
        t.Fatalf("read embedded preset: %v", err)
    }
    if len(data) == 0 {
        t.Fatal("embedded preset is empty")
    }
}
