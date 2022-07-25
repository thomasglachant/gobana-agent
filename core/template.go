package core

import (
	"bytes"
	"embed"
	"text/template"
)

var TemplateFs embed.FS

func TemplateToString(templates []string, vars map[string]interface{}) string {
	var tmpl *template.Template
	var err error
	var tpl bytes.Buffer

	tmpl, err = template.ParseFS(TemplateFs, templates...)
	if err != nil {
		Logger.Errorf("template", "unable to parse template : %s", err)
	}

	err = tmpl.Execute(&tpl, vars)
	if err != nil {
		Logger.Errorf("template", "unable to execute template : %s", err)
	}

	return tpl.String()
}

func CheckTemplateExists(filename string) bool {
	if _, err := TemplateFs.Open(filename); err != nil {
		return false
	}
	return true
}
