package presets

import "embed"

// FS contains embedded preset data for BootTree.
//
//go:embed */preset.json
var FS embed.FS
