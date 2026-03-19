package bcc

import (
	"log"
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

func (v *Vdc) GetKubernetesTemplates(extraArgs ...Arguments) (templates []*KubernetesTemplate, err error) {
	path := "/v1/kubernetes_template"
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)

	if err = v.manager.GetItems(path, args, &templates); err != nil {
		log.Printf("[REQUEST-ERROR] get-KubernetesTemplates failed: %s", err)
	} else {
		for i := range templates {
			templates[i].manager = v.manager
		}
	}

	return
}

func (m *Manager) GetKubernetesTemplate(id string) (template *KubernetesTemplate, err error) {
	path, _ := url.JoinPath("v1/kubernetes_template", id)

	if err = m.Get(path, Defaults(), &template); err != nil {
		log.Printf("[REQUEST-ERROR] get-KubernetesTemplate was failed: %s", err)
	} else {
		template.manager = m
	}

	return
}
