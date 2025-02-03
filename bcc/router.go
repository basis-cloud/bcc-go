package bcc

import (
	"fmt"
	"net/url"
)

type Router struct {
	manager   *Manager
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Vdc       struct {
		Id string `json:"id"`
	} `json:"vdc"`
	Ports    []*Port  `json:"ports"`
	Routes   []*Route `json:"routes"`
	Floating *Port    `json:"floating"`
	Locked   bool     `json:"locked"`
	Tags     []Tag    `json:"tags"`
}

func NewRouter(name string, floating *string) Router {
	r := Router{Name: name}
	if floating != nil {
		r.Floating = &Port{IpAddress: floating}
	}
	return r
}

func (m *Manager) GetRouters(extraArgs ...Arguments) (routers []*Router, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/router"
	err = m.GetItems(path, args, &routers)
	for i := range routers {
		routers[i].manager = m
		for x := range routers[i].Ports {
			routers[i].Ports[x].manager = m
		}
		for x := range routers[i].Routes {
			routers[i].Routes[x].router = routers[i]
		}
	}
	return
}

func (v *Vdc) GetRouters(extraArgs ...Arguments) (routers []*Router, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	routers, err = v.manager.GetRouters(args)
	return
}

func (m *Manager) GetRouter(id string) (router *Router, err error) {
	path, _ := url.JoinPath("v1/router", id)
	err = m.Get(path, Defaults(), &router)
	if err != nil {
		return
	}
	router.manager = m
	for _, port := range router.Ports {
		port.manager = m
	}
	for _, route := range router.Routes {
		route.router = router
	}
	return
}

func (r Router) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/router", r.ID)
	return loopWaitLock(r.manager, path)
}

func (r *Router) ConnectPort(port *Port, exsist bool) error {
	type TempPortCreate struct {
		Router      string   `json:"router"`
		Network     string   `json:"network"`
		IpAddress   *string  `json:"ip_address,omitempty"`
		FwTemplates []string `json:"fw_templates"`
	}

	var fwTemplates = make([]string, len(port.FirewallTemplates))
	for i, fwTemplate := range port.FirewallTemplates {
		fwTemplates[i] = fwTemplate.ID
	}
	args := &TempPortCreate{
		Router:      r.ID,
		Network:     port.Network.ID,
		IpAddress:   port.IpAddress,
		FwTemplates: fwTemplates,
	}

	var err error
	if exsist {
		path, _ := url.JoinPath("v1/port", port.ID)
		err = r.manager.Request("PUT", path, args, &port)

	} else {
		err = r.manager.Request("POST", "v1/port", args, &port)
	}

	if err == nil {
		port.manager = r.manager
	}

	return err
}

func (r *Router) DisconnectPort(port *Port) error {
	path := fmt.Sprintf("v1/port/%s/disconnect", port.ID)
	err := r.manager.Request("PATCH", path, Defaults(), &port)
	if err != nil {
		return err
	}
	for i, routerPorts := range r.Ports {
		if routerPorts == port {
			r.Ports = append(r.Ports[:i], r.Ports[i+1:]...)
			break
		}
	}

	return nil
}

func (r *Router) Delete() error {
	path, _ := url.JoinPath("v1/router", r.ID)
	return r.manager.Delete(path, Defaults(), nil)
}

func (r *Router) Rename(name string) error {
	path, _ := url.JoinPath("v1/router", r.ID)
	return r.manager.Request("PUT", path, Arguments{"name": name}, r.ID)
}

func (r *Router) Update() error {
	args := &struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IsDefault bool   `json:"is_default"`
		Vdc       struct {
			Id string `json:"id"`
		} `json:"vdc"`
		Ports    []*Port  `json:"ports"`
		Routes   []*Route `json:"routes"`
		Floating *string  `json:"floating"`
		Tags     []string `json:"tags"`
	}{
		ID:        r.ID,
		Name:      r.Name,
		IsDefault: r.IsDefault,
		Vdc:       r.Vdc,
		Ports:     r.Ports,
		Routes:    r.Routes,
		Tags:      convertTagsToNames(r.Tags),
	}
	if r.Floating == nil {
		args.Floating = nil
	} else {
		args.Floating = &r.Floating.ID
	}
	path, _ := url.JoinPath("v1/router", r.ID)
	r.WaitLock()
	return r.manager.Request("PUT", path, args, r)
}
