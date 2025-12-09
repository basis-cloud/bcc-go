package bcc

import (
	"fmt"
	"log"
	"net/url"

	"github.com/pkg/errors"
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
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/vdc"
	if err = m.GetItems(path, args, &vdcs); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return nil, errors.Wrapf(err, "Get vdcs %s failed", path)
	}
	for i := range vdcs {
		vdcs[i].manager = m
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
		log.Printf("[REQUEST-ERROR]: getting VDC-%s was failed: %s]", id, errors.WithStack(err))
		return nil, err
	} else {
		vdc.manager = m
	}
	return
}

func (p *Project) CreateVdc(vdc *Vdc) error {
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

	if err := p.manager.Request("POST", "v1/vdc", args, &vdc); err != nil {
		return err
	} else {
		vdc.manager = p.manager
	}

	return nil
}

func (v *Vdc) Rename(name string) error {
	v.Name = name
	return v.Update()
}

func (v *Vdc) Update() error {
	args := &struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		Name: v.Name,
		Tags: convertTagsToNames(v.Tags),
	}
	path, _ := url.JoinPath("v1/vdc", v.ID)
	return v.manager.Request("PUT", path, args, v)
}

func (v *Vdc) Delete() error {
	path, _ := url.JoinPath("v1/vdc", v.ID)
	if err := v.manager.Delete(path, Defaults(), nil); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return fmt.Errorf("request for delete vdc %s was failed", v.ID)
	}
	return nil
}

func (v Vdc) WaitLock() error {
	path, _ := url.JoinPath("v1/vdc", v.ID)
	if err := loopWaitLock(v.manager, path); err != nil {
		log.Printf("[REQUEST-ERROR]: %s]", errors.WithStack(err))
		return fmt.Errorf("request for WaitLock to VDC was failed")
	} else {
		return nil
	}
}
