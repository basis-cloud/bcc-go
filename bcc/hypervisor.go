package bcc

import (
	"log"
	"net/url"
)

type Hypervisor struct {
	manager        *Manager
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	CpuPerVm       int    `json:"cpu_per_vm"`
	RamPerVm       int    `json:"ram_per_vm"`
	PortsPerDevice int    `json:"ports_per_device"`
	DisksPerVm     int    `json:"disks_per_vm"`
}

func (p *Project) GetAvailableHypervisors(extraArgs ...Arguments) (hypervisors []*Hypervisor, err error) {
	path, _ := url.JoinPath("v1/project", p.ID)
	type tempType struct {
		Client struct {
			AllowedHypervisors []*Hypervisor `json:"allowed_hypervisors"`
		} `json:"client"`
	}

	var target tempType
	args := Defaults()
	args.merge(extraArgs)

	if err = p.manager.Get(path, args, &target); err != nil {
		log.Printf("[REQUEST-ERROR] get-projects for hypervisor was failed: %s", err)
	} else {
		hypervisors = target.Client.AllowedHypervisors

		for i := range hypervisors {
			hypervisors[i].manager = p.manager
		}
	}

	return
}
