package core

import (
	"bytes"
	"embed"
	"html/template"
	htmlTemplate "text/template"
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

func TemplateToHTML(templates []string, vars map[string]interface{}) string {
	var tmpl *htmlTemplate.Template
	var err error
	var tpl bytes.Buffer

	tmpl, err = htmlTemplate.ParseFS(TemplateFs, templates...)
	if err != nil {
		Logger.Errorf("template", "unable to parse template : %s", err)
	}

	err = tmpl.Execute(&tpl, vars)
	if err != nil {
		Logger.Errorf("template", "unable to execute template : %s", err)
	}

	return tpl.String()
}
