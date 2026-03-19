package bcc

import (
	"errors"
	"log"
	"net/url"
)

type Floating struct {
	ID        string `json:"id"`
	IpAddress string `json:"ip_address"`
}

func (m *Manager) GetFloating(id string) (fip *Floating, err error) {
	path, _ := url.JoinPath("v1/floating", id)

	if err = m.Get(path, Defaults(), &fip); err != nil {
		log.Printf("[REQUEST-ERROR] get-floating with id='%s' was failed: %s", id, err)
	}

	return
}

func (v *Vdc) GetFloatingByAddress(address string) (fip *Floating, err error) {
	path := "v1/port"
	args := Arguments{
		"vdc":         v.ID,
		"filter_type": "external",
	}
	var items []*Floating

	if err = v.manager.GetItems(path, args, &items); err != nil {
		log.Printf("[REQUEST-ERROR] get-floating by address '%s' was failed: %s", address, err)
	} else {
		for i := 0; i < len(items); i++ {
			if items[i].IpAddress == address {
				fip = items[i]
				return
			}
		}
	}

	return nil, errors.New("ERROR. Address not found")
}
