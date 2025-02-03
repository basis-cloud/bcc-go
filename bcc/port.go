package bcc

import (
	"fmt"
	"net/url"
)

type Port struct {
	manager           *Manager
	ID                string              `json:"id"`
	IpAddress         *string             `json:"ip_address,omitempty"`
	Network           *Network            `json:"network"`
	FirewallTemplates []*FirewallTemplate `json:"fw_templates,omitempty"`
	Connected         *Connected          `json:"connected"`
	Locked            bool                `json:"locked"`
	Tags              []Tag               `json:"tags"`
}

type Connected struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Vdc  *Vdc   `json:"vdc"`
}

func NewPort(network *Network, firewallTemplates []*FirewallTemplate, ipAddress string) Port {
	p := Port{Network: network, FirewallTemplates: firewallTemplates, IpAddress: &ipAddress}
	return p
}

func (v *Vdc) GetPorts(extraArgs ...Arguments) (ports []*Port, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	args.merge(extraArgs)

	path := "v1/port"
	err = v.manager.GetItems(path, args, &ports)
	for i := range ports {
		ports[i].manager = v.manager
		ports[i].Network.manager = v.manager
	}
	return
}

func (p *Port) UpdateFirewall(firewallTemplates []*FirewallTemplate) error {
	p.FirewallTemplates = firewallTemplates
	return p.Update()
}

func (p *Port) UpdateIpAddress(ip_address *string) error {
	p.IpAddress = ip_address
	return p.Update()
}

func (p *Port) Update() error {
	path, _ := url.JoinPath("v1/port", p.ID)
	fwTemplates := make([]*string, 0)
	for _, fwTemplate := range p.FirewallTemplates {
		fwTemplates = append(fwTemplates, &fwTemplate.ID)
	}
	args := &struct {
		IpAddress     *string   `json:"ip_address,omitempty"`
		FwTemplates   []*string `json:"fw_templates"`
		SecurityRules []string  `json:"security_rules"`
		Tags          []string  `json:"tags"`
	}{
		IpAddress:     p.IpAddress,
		FwTemplates:   fwTemplates,
		SecurityRules: []string{},
		Tags:          convertTagsToNames(p.Tags),
	}
	return p.manager.Request("PUT", path, args, p)
}

func (p *Port) Delete() error {
	path, _ := url.JoinPath("v1/port", p.ID)
	return p.manager.Delete(path, Defaults(), nil)
}

func (p *Port) ForceDelete() error {
	path := fmt.Sprintf("v1/port/%s/force", p.ID)
	return p.manager.Delete(path, Defaults(), nil)
}

func (r *Router) CreatePort(port *Port, toConnect interface{}) (err error) {
	args := &struct {
		manager           *Manager
		ID                string              `json:"id"`
		IpAddress         *string             `json:"ip_address,omitempty"`
		Network           string              `json:"network"`
		Router            string              `json:"router,omitempty"`
		Vm                string              `json:"vm,omitempty"`
		Lbaas             string              `json:"lbaas,omitempty"`
		FirewallTemplates []*FirewallTemplate `json:"fw_templates,omitempty"`
	}{
		ID:                port.ID,
		IpAddress:         port.IpAddress,
		Network:           port.Network.ID,
		FirewallTemplates: port.FirewallTemplates,
	}
	switch v := toConnect.(type) {
	case *Router:
		args.Router = v.ID
	case *Vm:
		args.Vm = v.ID
	default:
		return fmt.Errorf("ERROR. Unknown type: %s", v)
	}
	err = r.manager.Request("POST", "v1/port", args, &port)
	return
}

func (m *Manager) GetPort(id string) (port *Port, err error) {
	path, _ := url.JoinPath("v1/port", id)
	err = m.Get(path, Defaults(), &port)
	if err != nil {
		return
	}
	port.manager = m
	return
}

func (p Port) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/port", p.ID)
	return loopWaitLock(p.manager, path)
}
