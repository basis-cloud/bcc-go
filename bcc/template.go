package bcc

import (
	"log"
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

	if err = m.Get(path, Defaults(), &template); err != nil {
		log.Printf("[REQUEST-ERROR] get-template with id='%s' was failed: %s", id, err)
	} else {
		template.manager = m
	}

	return
}

func (v *Vdc) GetTemplates(extraArgs ...Arguments) (templates []*Template, err error) {
	path := "v1/template"
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)

	if err = v.manager.Get(path, args, &templates); err != nil {
		log.Printf("[REQUEST-ERROR] get-templates was failed: %s", err)
	} else {
		for i := range templates {
			templates[i].manager = v.manager
		}
	}

	return
}
