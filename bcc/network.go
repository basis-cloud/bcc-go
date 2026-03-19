package bcc

import (
	"fmt"
	"log"
	"net/url"

	"github.com/pkg/errors"
)

type Network struct {
	manager   *Manager
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	External  bool   `json:"external"`
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
	path := "v1/network"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &networks); err != nil {
		log.Printf("[REQUEST-ERROR]: get-Network list failed: %s]", err)
	} else {
		for i := range networks {
			networks[i].manager = m
		}
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

	if err = m.Get(path, Defaults(), &network); err != nil {
		log.Printf("[REQUEST-ERROR]: get-network with id='%s' failed: %s", id, err)
	} else {
		network.manager = m
		for i := range network.Subnets {
			network.Subnets[i].network = network
			network.Subnets[i].manager = m
		}
	}

	return
}

func (v *Vdc) CreateNetwork(network *Network) error {
	path := "v1/network"
	args := &struct {
		Name string   `json:"name"`
		Vdc  string   `json:"vdc"`
		Mtu  *int     `json:"mtu,omitempty"`
		Tags []string `json:"tags"`
	}{
		Name: network.Name,
		Vdc:  v.ID,
		Mtu:  network.Mtu,
		Tags: convertTagsToNames(network.Tags),
	}

	if err := v.manager.Request("POST", path, args, &network); err != nil {
		log.Printf("[REQUEST-ERROR]: create-network failed: %s", err)
	} else {
		network.manager = v.manager
	}

	return nil
}

func (n *Network) GetSubnets(extraArgs ...Arguments) (subnets []*Subnet, err error) {
	args := Defaults()
	args.merge(extraArgs)
	path := fmt.Sprintf("v1/network/%s/subnet", n.ID)
	if err = n.manager.GetItems(path, args, &subnets); err != nil {
		return subnets, errors.Wrapf(err, "[REQUEST-ERROR]: get-Subnets with network id='%s' failed: %s", n.ID, err)
	}
	for i := range subnets {
		subnets[i].manager = n.manager
		subnets[i].network = n
	}

	return
}

func (n *Network) CreateSubnet(subnet *Subnet) (err error) {
	path := fmt.Sprintf("v1/network/%s/subnet", n.ID)

	if err = n.manager.Request("POST", path, subnet, &subnet); err != nil {
		return errors.Wrapf(err, "[REQUEST-ERROR]: create-Subnet failed: %s", err)
	} else {
		subnet.manager = n.manager
		subnet.network = n
	}

	return
}

func (n *Network) Rename(name string) error {
	n.Name = name
	return n.Update()
}

func (n *Network) Update() (err error) {
	path, _ := url.JoinPath("v1/network", n.ID)
	args := &struct {
		Name string   `json:"name"`
		Mtu  *int     `json:"mtu,omitempty"`
		Tags []string `json:"tags"`
	}{
		Name: n.Name,
		Mtu:  n.Mtu,
		Tags: convertTagsToNames(n.Tags),
	}

	if err = n.manager.Request("PUT", path, args, n); err != nil {
		log.Printf("[REQUEST-ERROR]: update-network failed: %s", err)
	}

	return
}

func (n *Network) Delete() (err error) {
	path, _ := url.JoinPath("v1/network", n.ID)
	if err = n.manager.Delete(path, Defaults(), nil); err != nil {
		log.Printf("[REQUEST-ERROR]: delete-network failed: %s", err)
	}

	return
}

func (n Network) WaitLock() error {
	path, _ := url.JoinPath("v1/network", n.ID)
	return loopWaitLock(n.manager, path)
}
