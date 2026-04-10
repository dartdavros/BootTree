package templates

import "embed"

// FS contains embedded template data for BootTree.
//
//go:embed **/*
var FS embed.FS
