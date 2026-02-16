package bcc

import (
	"net/url"
)

type Hypervisor struct {
	manager  *Manager
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	CpuPerVm int    `json:"cpu_per_vm"`
	RamPerVm int    `json:"ram_per_vm"`
}

func (p *Project) GetAvailableHypervisors(extraArgs ...Arguments) (hypervisors []*Hypervisor, err error) {
	type tempType struct {
		Client struct {
			AllowedHypervisors []*Hypervisor `json:"allowed_hypervisors"`
		} `json:"client"`
	}

	var target tempType
	args := Defaults()
	args.merge(extraArgs)

	path, _ := url.JoinPath("v1/project", p.ID)
	err = p.manager.Get(path, args, &target)
	hypervisors = target.Client.AllowedHypervisors

	for i := range hypervisors {
		hypervisors[i].manager = p.manager
	}
	return
}
