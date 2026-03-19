package bcc

import (
	"log"
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
	path := "v1/dns"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &dnss); err != nil {
		log.Printf("[REQUEST-ERROR] get-dns's was failed: %s", err)
	} else {
		for i := range dnss {
			dnss[i].manager = m
		}
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

	if err = m.Get(path, Defaults(), &dns); err != nil {
		log.Printf("[REQUEST-ERROR] get-dns with id='%s' was failed: %s", id, err)
	} else {
		dns.manager = m
	}

	return
}

func (p *Project) CreateDns(dns *Dns) (err error) {
	path := "v1/dns"
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

	if err = p.manager.Request("POST", path, args, &dns); err != nil {
		log.Printf("[REQUEST-ERROR] create-dns failed: %s", err)
	} else {
		dns.manager = p.manager
	}

	return
}

func (d *Dns) Delete() error {
	path, _ := url.JoinPath("v1/dns", d.ID)
	return d.manager.Delete(path, Defaults(), nil)
}

func (d *Dns) Update() (err error) {
	path, _ := url.JoinPath("v1/dns", d.ID)

	args := &struct {
		Name    string   `json:"name"`
		Project string   `json:"project"`
		Tags    []string `json:"tags"`
	}{
		Name:    d.Name,
		Project: d.Project.ID,
		Tags:    convertTagsToNames(d.Tags),
	}

	if err := d.manager.Request("PUT", path, args, d); err != nil {
		log.Printf("[REQUEST-ERROR] update-dns failed: %s", err)
	}

	return
}
