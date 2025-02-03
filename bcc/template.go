package bcc

import (
	"net/url"
)

type Template struct {
	manager *Manager
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	MinCpu  int     `json:"min_cpu"`
	MinRam  float64 `json:"min_ram"`
	MinHdd  int     `json:"min_hdd"`
}

func (m *Manager) GetTemplate(id string) (template *Template, err error) {
	path, _ := url.JoinPath("v1/template", id)
	err = m.Get(path, Defaults(), &template)
	if err != nil {
		return
	}
	template.manager = m
	return
}

func (v *Vdc) GetTemplates() (templates []*Template, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	path := "v1/template"
	err = v.manager.Get(path, args, &templates)
	for i := range templates {
		templates[i].manager = v.manager
	}
	return
}
