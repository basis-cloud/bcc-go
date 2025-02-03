package bcc

import (
	"net/url"
)

type Hypervisor struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

func (p *Project) GetAvailableHypervisors() (hypervisors []*Hypervisor, err error) {
	type tempType struct {
		Client struct {
			AllowedHypervisors []*Hypervisor `json:"allowed_hypervisors"`
		} `json:"client"`
	}

	var target tempType

	path, _ := url.JoinPath("v1/project", p.ID)
	err = p.manager.Get(path, Defaults(), &target)
	hypervisors = target.Client.AllowedHypervisors

	for i := range hypervisors {
		hypervisors[i].manager = p.manager
	}
	return
}
