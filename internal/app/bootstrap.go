package app

import (
	"boottree/internal/core/preset"
	coretemplate "boottree/internal/core/template"
)

type Bootstrap struct {
	Presets   preset.EmbeddedRepository
	Templates coretemplate.EmbeddedRepository
	Renderer  coretemplate.Renderer
}

func NewBootstrap() Bootstrap {
	return Bootstrap{
		Presets:   preset.NewEmbeddedRepository(),
		Templates: coretemplate.NewEmbeddedRepository(),
		Renderer:  coretemplate.Renderer{},
	}
}
