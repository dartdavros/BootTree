package assets

import "embed"

//go:embed presets/*/preset.json templates/**/*
var FS embed.FS
