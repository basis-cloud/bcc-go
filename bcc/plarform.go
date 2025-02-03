package bcc

import (
	"net/url"
)

type Platform struct {
	manager    *Manager
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Hypervisor *Hypervisor `json:"hypervisor"`
}

func (m *Manager) GetPlatforms(vdc_id string) (platforms []*Platform, err error) {
	args := Arguments{
		"vdc": vdc_id,
	}

	path := "v1/platform"
	err = m.Get(path, args, &platforms)
	if err != nil {
		return
	}
	for i := range platforms {
		platforms[i].manager = m
	}
	return
}

func (m *Manager) GetPlatform(id string) (platforms *Platform, err error) {
	path, _ := url.JoinPath("v1/platform", id)
	err = m.Get(path, Defaults(), &platforms)
	if err != nil {
		return
	}
	platforms.manager = m
	return
}
