package mail

import (
	"bytes"
	"html/template"
	"strings"
)

type TemplateEngine interface {
	Render(content string, data any) (string, error)
}

type GoTemplateEngine struct {
	EnableHTML bool
}

func (g *GoTemplateEngine) Render(content string, data any) (string, error) {
	if !strings.Contains(content, "{{") {
		return content, nil
	}
	var (
		tmpl *template.Template
		err  error
	)
	if g.EnableHTML {
		tmpl, err = template.New("mail").Parse(content)
	} else {
		// 使用 text/template（不会进行 HTML 转义）
		tmpl, err = template.New("mail").Option("missingkey=error").Parse(content)
	}
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
