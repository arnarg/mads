package envoy

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed tpl/*
var tpl embed.FS

type TemplateParams struct {
	AdminAddress string
	AdminPort    uint16
	ServiceName  string
	ServiceID    string
	AgentAddress string
	AgentPort    uint16
	AgentTLS     bool
	AgentCAPEM   string
	ConsulToken  string
}

func TemplateConfig(params *TemplateParams) (string, error) {
	// Parse template from embedded file
	templ, err := template.ParseFS(tpl, "tpl/envoy.yml.tpl")
	if err != nil {
		return "", err
	}

	// Render template into buffer
	buf := &bytes.Buffer{}
	err = templ.Execute(buf, params)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
