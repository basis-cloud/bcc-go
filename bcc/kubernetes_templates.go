package bcc

import (
	"net/url"
)

type KubernetesTemplate struct {
	manager    *Manager
	ID         string `json:"id"`
	Name       string `json:"name"`
	MinNodeCpu int    `json:"min_node_cpu"`
	MinNodeRam int    `json:"min_node_ram"`
	MinNodeHdd int    `json:"min_node_hdd"`
}

func (v *Vdc) GetKubernetesTemplates() (templates []*KubernetesTemplate, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	path := "/v1/kubernetes_template"
	err = v.manager.GetItems(path, args, &templates)
	for i := range templates {
		templates[i].manager = v.manager
	}
	return
}

func (m *Manager) GetKubernetesTemplate(id string) (template *KubernetesTemplate, err error) {
	path, _ := url.JoinPath("v1/kubernetes_template", id)
	err = m.Get(path, Defaults(), &template)
	if err != nil {
		return
	}
	template.manager = m
	return
}
