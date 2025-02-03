package bcc

import (
	"fmt"
	"net/url"
)

type Network struct {
	manager   *Manager
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Mtu       *int   `json:"mtu,omitempty"`
	Vdc       struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"vdc"`
	Locked  bool     `json:"locked"`
	Subnets []Subnet `json:"subnets"`
	Tags    []Tag    `json:"tags"`
}

func NewNetwork(name string) Network {
	n := Network{Name: name}
	return n
}

func (m *Manager) GetNetworks(extraArgs ...Arguments) (networks []*Network, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/network"
	err = m.GetItems(path, args, &networks)
	for i := range networks {
		networks[i].manager = m
	}
	return
}

func (v *Vdc) GetNetworks(extraArgs ...Arguments) (networks []*Network, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	args.merge(extraArgs)
	networks, err = v.manager.GetNetworks(args)
	return
}

func (m *Manager) GetNetwork(id string) (network *Network, err error) {
	path := fmt.Sprintf("v1/network/%s", id)
	err = m.Get(path, Defaults(), &network)
	if err != nil {
		return
	}
	network.manager = m
	for i := range network.Subnets {
		network.Subnets[i].network = network
		network.Subnets[i].manager = m
	}
	return
}

func (n *Network) CreateSubnet(subnet *Subnet) error {
	path := fmt.Sprintf("v1/network/%s/subnet", n.ID)
	err := n.manager.Request("POST", path, subnet, &subnet)
	if err == nil {
		subnet.manager = n.manager
		subnet.network = n
	}

	return err
}

func (n *Network) Rename(name string) error {
	n.Name = name
	return n.Update()
}

func (n *Network) Update() error {
	args := &struct {
		Name string   `json:"name"`
		Mtu  *int     `json:"mtu,omitempty"`
		Tags []string `json:"tags"`
	}{
		Name: n.Name,
		Mtu:  n.Mtu,
		Tags: convertTagsToNames(n.Tags),
	}
	path, _ := url.JoinPath("v1/network", n.ID)
	return n.manager.Request("PUT", path, args, n)
}

func (n *Network) GetSubnets() (subnets []*Subnet, err error) {
	path := fmt.Sprintf("v1/network/%s/subnet", n.ID)
	err = n.manager.GetItems(path, Arguments{}, &subnets)
	for i := range subnets {
		subnets[i].manager = n.manager
		subnets[i].network = n
	}

	return
}

func (n *Network) Delete() error {
	path, _ := url.JoinPath("v1/network", n.ID)
	return n.manager.Delete(path, Defaults(), nil)
}

func (n Network) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/network", n.ID)
	return loopWaitLock(n.manager, path)
}
