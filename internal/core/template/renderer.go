package template

import (
	"bytes"
	texttemplate "text/template"

	"boottree/internal/core/model"
)

type Renderer struct{}

func (Renderer) Render(templateText string, data model.TemplateData) (string, error) {
	tpl, err := texttemplate.New("template").Option("missingkey=error").Parse(templateText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
