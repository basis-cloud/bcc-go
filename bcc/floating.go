package bcc

import (
	"fmt"
	"net/url"
)

type Floating struct {
	ID        string `json:"id"`
	IpAddress string `json:"ip_address"`
}

func (m *Manager) GetFloating(id string) (fip *Floating, err error) {
	path, _ := url.JoinPath("v1/floating", id)
	err = m.Get(path, Defaults(), &fip)
	return
}

func (v *Vdc) GetFloatingByAddress(address string) (fip *Floating, err error) {
	args := Arguments{
		"vdc":         v.ID,
		"filter_type": "external",
	}
	var items []*Floating
	err = v.manager.GetItems("v1/port", args, &items)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(items); i++ {
		if items[i].IpAddress == address {
			fip = items[i]
			return
		}
	}
	return nil, fmt.Errorf("ERROR. Address %s not found", address)
}
