package bcc

import (
	"net/url"
)

type Vdc struct {
	manager    *Manager
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Locked     bool       `json:"locked"`
	Hypervisor Hypervisor `json:"hypervisor"`

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
	err = m.GetItems(path, args, &vdcs)
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
	err = m.Get(path, Defaults(), &vdc)
	if err != nil {
		return
	}
	vdc.manager = m
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

	err := p.manager.Request("POST", "v1/vdc", args, &vdc)
	if err == nil {
		vdc.manager = p.manager
	}

	return err
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
	return v.manager.Delete(path, Defaults(), nil)
}

func (v *Vdc) CreateNetwork(network *Network) error {
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

	err := v.manager.Request("POST", "v1/network", args, &network)
	if err == nil {
		network.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateRouter(router *Router, ports ...*Port) error {
	type TempPortCreate struct {
		ID string `json:"id"`
	}

	tempPorts := make([]*TempPortCreate, len(ports))
	for idx := range ports {
		tempPorts[idx] = &TempPortCreate{ID: ports[idx].ID}
	}

	args := &struct {
		Name     string            `json:"name"`
		Vdc      string            `json:"vdc"`
		Ports    []*TempPortCreate `json:"ports"`
		Routes   []*Route          `json:"routes"`
		Floating *string           `json:"floating"`
		Tags     []string          `json:"tags"`
	}{
		Name:     router.Name,
		Vdc:      v.ID,
		Ports:    tempPorts,
		Routes:   router.Routes,
		Floating: nil,
		Tags:     convertTagsToNames(router.Tags),
	}

	if router.Floating != nil {
		if router.Floating.ID != "" {
			args.Floating = &router.Floating.ID
		} else {
			args.Floating = router.Floating.IpAddress
		}
	}

	err := v.manager.Request("POST", "v1/router", args, &router)
	if err == nil {
		router.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateVm(vm *Vm) error {
	type TempPortCreate struct {
		ID string `json:"id"`
	}

	tempPorts := make([]*TempPortCreate, len(vm.Ports))
	for idx := range vm.Ports {
		tempPorts[idx] = &TempPortCreate{ID: vm.Ports[idx].ID}
	}

	type TempFields struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}

	tempFields := make([]*TempFields, len(vm.Metadata))
	for idx := range vm.Metadata {
		tempFields[idx] = &TempFields{Field: vm.Metadata[idx].Field.ID, Value: vm.Metadata[idx].Value}
	}

	type TempDisk struct {
		Name           string `json:"name"`
		Size           int    `json:"size"`
		StorageProfile string `json:"storage_profile"`
	}

	tempDisks := make([]*TempDisk, len(vm.Disks))
	for idx := range vm.Disks {
		tempDisks[idx] = &TempDisk{Name: vm.Disks[idx].Name, Size: vm.Disks[idx].Size, StorageProfile: vm.Disks[idx].StorageProfile.ID}
	}

	args := &struct {
		Name     string            `json:"name"`
		Cpu      int               `json:"cpu"`
		Ram      float64           `json:"ram"`
		Vdc      string            `json:"vdc"`
		Template string            `json:"template"`
		Ports    []*TempPortCreate `json:"ports"`
		Metadata []*TempFields     `json:"metadata"`
		UserData *string           `json:"user_data,omitempty"`
		Disks    []*TempDisk       `json:"disks"`
		Floating *string           `json:"floating"`
		Tags     []string          `json:"tags"`
	}{
		Name:     vm.Name,
		Cpu:      vm.Cpu,
		Ram:      vm.Ram,
		Vdc:      v.ID,
		Template: vm.Template.ID,
		Ports:    tempPorts,
		Metadata: tempFields,
		UserData: vm.UserData,
		Disks:    tempDisks,
		Floating: nil,
		Tags:     convertTagsToNames(vm.Tags),
	}

	if vm.Floating != nil {
		if vm.Floating.ID != "" {
			args.Floating = &vm.Floating.ID
		} else {
			args.Floating = vm.Floating.IpAddress
		}
	}

	err := v.manager.Request("POST", "v1/vm", args, &vm)
	if err == nil {
		vm.manager = v.manager
		for idx := range vm.Ports {
			vm.Ports[idx].manager = v.manager
		}
		for idx := range vm.Disks {
			vm.Disks[idx].manager = v.manager
		}
		if vm.Floating != nil {
			vm.Floating.manager = v.manager
		}
	}

	return err
}

func (v *Vdc) CreateKubernetes(k *Kubernetes) error {
	type TempPortCreate struct {
		ID string `json:"id"`
	}

	args := &struct {
		Name               string   `json:"name"`
		NodeCpu            int      `json:"node_cpu"`
		NodeRam            int      `json:"node_ram"`
		NodeDiskSize       int      `json:"node_disk_size"`
		NodesCount         int      `json:"nodes_count"`
		NodeStorageProfile *string  `json:"node_storage_profile"`
		Vdc                *string  `json:"vdc"`
		Template           *string  `json:"template"`
		Floating           *string  `json:"floating"`
		UserPublicKey      string   `json:"user_public_key"`
		NodePlatform       string   `json:"node_platform"`
		Tags               []string `json:"tags"`
	}{
		Name:               k.Name,
		NodeCpu:            k.NodeCpu,
		NodeRam:            k.NodeRam,
		NodeDiskSize:       k.NodeDiskSize,
		NodesCount:         k.NodesCount,
		NodeStorageProfile: &k.NodeStorageProfile.ID,
		Vdc:                &v.ID,
		Template:           &k.Template.ID,
		UserPublicKey:      k.UserPublicKey,
		Floating:           nil,
		NodePlatform:       k.NodePlatform.ID,
		Tags:               convertTagsToNames(k.Tags),
	}

	if k.Floating != nil {
		args.Floating = k.Floating.IpAddress
	}

	err := v.manager.Request("POST", "/v1/kubernetes", args, &k)
	if err == nil {
		k.manager = v.manager
		for idx := range k.Vms {
			k.Vms[idx].manager = v.manager
		}

	}

	return err
}

func (v *Vdc) CreateDisk(disk *Disk) error {
	args := &struct {
		Name           string   `json:"name"`
		Vdc            *string  `json:"vdc,omitempty"`
		Vm             *string  `json:"vm,omitempty"`
		Size           int      `json:"size"`
		StorageProfile string   `json:"storage_profile"`
		Tags           []string `json:"tags"`
	}{
		Name:           disk.Name,
		Vdc:            &v.ID,
		Vm:             nil,
		Size:           disk.Size,
		StorageProfile: disk.StorageProfile.ID,
		Tags:           convertTagsToNames(disk.Tags),
	}

	if disk.Vm != nil {
		args.Vm = &disk.Vm.ID
		args.Vdc = nil
	}

	err := v.manager.Request("POST", "v1/disk", args, &disk)
	if err == nil {
		disk.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateEmptyPort(port *Port) (err error) {
	var fwTemplates = make([]*string, 0)
	for _, fwTemplate := range port.FirewallTemplates {
		fwTemplates = append(fwTemplates, &fwTemplate.ID)
	}
	args := &struct {
		manager     *Manager
		ID          string    `json:"id"`
		IpAddress   *string   `json:"ip_address,omitempty"`
		Network     string    `json:"network"`
		FwTemplates []*string `json:"fw_templates"`
		Tags        []string  `json:"tags"`
	}{
		ID:          port.ID,
		IpAddress:   port.IpAddress,
		Network:     port.Network.ID,
		FwTemplates: fwTemplates,
		Tags:        convertTagsToNames(port.Tags),
	}

	err = v.manager.Request("POST", "v1/port", args, &port)
	if err == nil {
		port.manager = v.manager
	}

	return
}

func (v Vdc) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/vdc", v.ID)
	return loopWaitLock(v.manager, path)
}

func (v Vdc) Create(lb *LoadBalancer) (err error) {
	type customPort struct {
		ID                string     `json:"id"`
		IpAddress         *string    `json:"ip_address,omitempty"`
		Network           string     `json:"network"`
		FirewallTemplates *string    `json:"fw_templates,omitempty"`
		Connected         *Connected `json:"connected"`
	}
	lbCreate := &struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Vdc  string `json:"vdc"`

		Kubernetes *Kubernetes `json:"kubernetes"`
		Port       customPort  `json:"port"`
		Floating   *string     `json:"floating"`
		Tags       []string    `json:"tags"`
	}{
		Name: lb.Name,
		Vdc:  lb.Vdc.ID,
		Port: customPort{
			ID:                lb.Port.ID,
			IpAddress:         lb.Port.IpAddress,
			Network:           lb.Port.Network.ID,
			FirewallTemplates: nil,
			Connected:         lb.Port.Connected,
		},
		Kubernetes: lb.Kubernetes,
		Floating:   nil,
		Tags:       convertTagsToNames(lb.Tags),
	}
	if lb.Floating != nil {
		lbCreate.Floating = lb.Floating.IpAddress
	}
	err = lb.manager.Request("POST", "v1/lbaas", lbCreate, &lb)
	if err == nil {
		lb.manager = v.manager
	}
	return
}
