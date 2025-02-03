package bcc

import (
	"net/url"
)

type Dns struct {
	manager *Manager
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Project *Project `json:"project"`
	Tags    []Tag    `json:"tags"`
}

func NewDns(name string) Dns {
	d := Dns{Name: name}
	return d
}

func (m *Manager) GetDnss(extraArgs ...Arguments) (dnss []*Dns, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/dns"
	err = m.GetItems(path, args, &dnss)
	for i := range dnss {
		dnss[i].manager = m
	}
	return
}

func (p *Project) GetDnss(extraArgs ...Arguments) (dns []*Dns, err error) {
	args := Arguments{
		"project": p.ID,
	}

	args.merge(extraArgs)
	dns, err = p.manager.GetDnss(args)
	return
}

func (m *Manager) GetDns(id string) (dns *Dns, err error) {
	path, _ := url.JoinPath("v1/dns", id)
	err = m.Get(path, Defaults(), &dns)
	if err != nil {
		return
	}
	dns.manager = m
	return
}

func (p *Project) CreateDns(dns *Dns) (err error) {
	args := &struct {
		manager *Manager
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Project string   `json:"project"`
		Tags    []string `json:"tags"`
	}{
		ID:      dns.ID,
		Name:    dns.Name,
		Project: p.ID,
		Tags:    convertTagsToNames(dns.Tags),
	}

	err = p.manager.Request("POST", "v1/dns", args, &dns)
	dns.manager = p.manager
	return
}

func (d *Dns) Delete() error {
	path, _ := url.JoinPath("v1/dns", d.ID)
	return d.manager.Delete(path, Defaults(), nil)
}
