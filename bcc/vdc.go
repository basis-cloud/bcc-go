package bcc

import (
	"log"
	"net/url"
)

type Vdc struct {
	manager    *Manager
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Locked     bool       `json:"locked"`
	Hypervisor Hypervisor `json:"hypervisor"`

	Paas *struct {
		ID     string `json:"id"`
		Locked bool   `json:"locked"`
	} `json:"paas"`

	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`

	Tags []Tag `json:"tags"`
}

func NewVdc(name string, hypervisor *Hypervisor) Vdc {
	v := Vdc{Name: name, Hypervisor: Hypervisor{ID: hypervisor.ID}}
	return v
}

func (m *Manager) GetVdcs(extraArgs ...Arguments) (vdcs []*Vdc, err error) {
	path := "v1/vdc"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &vdcs); err != nil {
		log.Printf("[REQUEST-ERROR] get-vdcs was failed: %s", err)
	} else {
		for i := range vdcs {
			vdcs[i].manager = m
		}
	}

	return
}

func (v *Vdc) GetVdcs(extraArgs ...Arguments) (vdcs []*Vdc, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	vdcs, err = v.manager.GetVdcs(args)
	return
}

func (m *Manager) GetVdc(id string) (vdc *Vdc, err error) {
	path, _ := url.JoinPath("v1/vdc", id)

	if err = m.Get(path, Defaults(), &vdc); err != nil {
		log.Printf("[REQUEST-ERROR] get-vdc with id='%s' was failed: %s", id, err)
	} else {
		vdc.manager = m
	}

	return
}

func (p *Project) CreateVdc(vdc *Vdc) (err error) {
	path := "v1/vdc"
	args := &struct {
		Name       string   `json:"name"`
		Hypervisor string   `json:"hypervisor"`
		Project    string   `json:"project"`
		Tags       []string `json:"tags"`
	}{
		Name:       vdc.Name,
		Hypervisor: vdc.Hypervisor.ID,
		Project:    p.ID,
		Tags:       convertTagsToNames(vdc.Tags),
	}

	if err = p.manager.Request("POST", path, args, &vdc); err != nil {
		log.Printf("[REQUEST-ERROR] create-vdc was failed: %s", err)
	} else {
		vdc.manager = p.manager
	}

	return
}

func (v *Vdc) Rename(name string) error {
	v.Name = name
	return v.Update()
}

func (v *Vdc) Update() (err error) {
	path, _ := url.JoinPath("v1/vdc", v.ID)
	args := &struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		Name: v.Name,
		Tags: convertTagsToNames(v.Tags),
	}

	if err = v.manager.Request("PUT", path, args, v); err != nil {
		log.Printf("[REQUEST-ERROR] update-vdc was failed: %s", err)
	}

	return
}

func (v *Vdc) Delete() (err error) {
	path, _ := url.JoinPath("v1/vdc", v.ID)

	if err = v.manager.Delete(path, Defaults(), nil); err != nil {
		log.Printf("[REQUEST-ERROR] delete-vdc was failed: %s", err)
	}

	return
}

func (v Vdc) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/vdc", v.ID)

	if err = loopWaitLock(v.manager, path); err != nil {
		log.Printf("[REQUEST-ERROR] wait-lock for vdc-%s was failed: %s", v.ID, err)
	}

	return
}
