package bcc

import (
	"log"
	"net/url"
)

type Platform struct {
	manager    *Manager
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Hypervisor *Hypervisor `json:"hypervisor"`
}

func (m *Manager) GetPlatforms(vdc_id string, extraArgs ...Arguments) (platforms []*Platform, err error) {
	path := "v1/platform"
	args := Arguments{
		"vdc": vdc_id,
	}
	args.merge(extraArgs)

	if err = m.Get(path, args, &platforms); err != nil {
		log.Printf("[REQUEST-ERROR]: get-platforms was failed: %s", err)
	} else {
		for i := range platforms {
			platforms[i].manager = m
		}
	}

	return
}

func (m *Manager) GetPlatform(id string) (platforms *Platform, err error) {
	path, _ := url.JoinPath("v1/platform", id)

	if err = m.Get(path, Defaults(), &platforms); err != nil {
		log.Printf("[REQUEST-ERROR]: get-platform was failed: %s", err)
	} else {
		platforms.manager = m
	}

	return
}
