package prompt

import (
	"bytes"
	"text/template"
)

// Template holds a prompt string with variable placeholders.
// Variables use {{.VarName}} syntax.
type Template struct {
	Raw       string
	Variables map[string]string
}

// Render substitutes variables into template.
func (t *Template) Render() (string, error) {
	tmpl, err := template.New("prompt").Option("missingkey=error").Parse(t.Raw)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, t.Variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}
