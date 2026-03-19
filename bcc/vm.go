package bcc

import (
	"fmt"
	"log"
	"net/url"
)

type Vm struct {
	manager        *Manager
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Cpu            int              `json:"cpu"`
	Ram            float64          `json:"ram"`
	Power          bool             `json:"power"`
	Vdc            *Vdc             `json:"vdc"`
	HotAdd         bool             `json:"hotadd_feature"`
	Template       *Template        `json:"template"`
	Metadata       []*VmMetadata    `json:"metadata"`
	UserData       *string          `json:"user_data"`
	Ports          []*Port          `json:"ports"`
	Disks          []*Disk          `json:"disks"`
	Floating       *Port            `json:"floating"`
	Locked         bool             `json:"locked,omitempty"`
	Platform       *Platform        `json:"platform,omitempty"`
	Tags           []Tag            `json:"tags"`
	Kubernetes     *MetaData        `json:"kubernetes,omitempty"`
	AffinityGroups []*AffinityGroup `json:"affinity_groups,omitempty"`
}

func NewVm(name string, cpu int, ram float64, template *Template, metadata []*VmMetadata, userData *string, ports []*Port, disks []*Disk, floating *string) Vm {
	v := Vm{Name: name, Cpu: cpu, Ram: ram, Power: true, Template: template, Metadata: metadata, UserData: userData, Ports: ports, Disks: disks}
	if floating != nil {
		v.Floating = &Port{IpAddress: floating}
	}
	return v
}

func (m *Manager) GetVms(extraArgs ...Arguments) (vms []*Vm, err error) {
	path := "v1/vm"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &vms); err != nil {
		log.Printf("[REQUEST-ERROR] get-vms was failed: %s", err)
	} else {
		for i := range vms {
			vms[i].manager = m
			for x := range vms[i].Ports {
				vms[i].Ports[x].manager = m
			}
			for x := range vms[i].Disks {
				vms[i].Disks[x].manager = m
			}
			vms[i].Vdc.manager = m
			if vms[i].Floating != nil {
				vms[i].Floating.manager = m
			}
		}
	}

	return
}

func (v *Vdc) GetVms(extraArgs ...Arguments) (vms []*Vm, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	vms, err = v.manager.GetVms(args)
	return
}

func (m *Manager) GetVm(id string) (vm *Vm, err error) {
	path, _ := url.JoinPath("v1/vm", id)

	if err = m.Get(path, Defaults(), &vm); err != nil {
		log.Printf("[REQUEST-ERROR] get-vm was failed: %s", err)
	} else {
		vm.manager = m
		for x := range vm.Ports {
			vm.Ports[x].manager = m
		}
		for x := range vm.Disks {
			vm.Disks[x].manager = m
		}
		vm.Vdc.manager = m
		if vm.Floating != nil {
			vm.Floating.manager = m
		}
	}

	return
}

func (v *Vdc) CreateVm(vm *Vm) (err error) {
	path := "v1/vm"
	type idList struct {
		ID string `json:"id"`
	}

	portList := make([]*idList, len(vm.Ports))
	for idx := range vm.Ports {
		portList[idx] = &idList{ID: vm.Ports[idx].ID}
	}

	type metadata struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}

	metaDataList := make([]*metadata, len(vm.Metadata))
	for idx := range vm.Metadata {
		metaDataList[idx] = &metadata{Field: vm.Metadata[idx].Field.ID, Value: vm.Metadata[idx].Value}
	}

	type TempDisk struct {
		Name           string `json:"name"`
		Size           int    `json:"size"`
		StorageProfile string `json:"storage_profile"`
	}

	diskList := make([]*TempDisk, len(vm.Disks))
	for idx := range vm.Disks {
		diskList[idx] = &TempDisk{
			Name:           vm.Disks[idx].Name,
			Size:           vm.Disks[idx].Size,
			StorageProfile: vm.Disks[idx].StorageProfile.ID,
		}
	}

	var affGrList []string
	if vm.AffinityGroups != nil && len(vm.AffinityGroups) > 0 {
		for _, group := range vm.AffinityGroups {
			affGrList = append(affGrList, group.ID)
		}
	}

	args := &struct {
		Name           string      `json:"name"`
		Cpu            int         `json:"cpu"`
		Ram            float64     `json:"ram"`
		Vdc            string      `json:"vdc"`
		Template       string      `json:"template"`
		HotAdd         bool        `json:"hotadd_feature"`
		Ports          []*idList   `json:"ports"`
		Metadata       []*metadata `json:"metadata"`
		UserData       *string     `json:"user_data,omitempty"`
		Disks          []*TempDisk `json:"disks"`
		Floating       *string     `json:"floating"`
		Tags           []string    `json:"tags"`
		Platform       *string     `json:"platform,omitempty"`
		AffinityGroups []string    `json:"affinity_groups,omitempty"`
	}{
		Name:           vm.Name,
		Cpu:            vm.Cpu,
		Ram:            vm.Ram,
		Vdc:            v.ID,
		Template:       vm.Template.ID,
		HotAdd:         vm.HotAdd,
		Ports:          portList,
		Metadata:       metaDataList,
		UserData:       vm.UserData,
		Disks:          diskList,
		Floating:       nil,
		Tags:           convertTagsToNames(vm.Tags),
		Platform:       nil,
		AffinityGroups: affGrList,
	}

	if vm.Floating != nil {
		if vm.Floating.ID != "" {
			args.Floating = &vm.Floating.ID
		} else {
			args.Floating = vm.Floating.IpAddress
		}
	}

	if vm.Platform != nil {
		args.Platform = &vm.Platform.ID
	}

	if err = v.manager.Request("POST", path, args, &vm); err != nil {
		log.Printf("[REQUEST-ERROR] create-vm was failed: %s", err)
	} else {
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

	return
}

func (v *Vm) ConnectPort(port *Port, exsist bool) (err error) {
	path := "v1/port"
	method := "POST"
	type TempPortCreate struct {
		Vm          string   `json:"vm"`
		Network     string   `json:"network"`
		IpAddress   *string  `json:"ip_address,omitempty"`
		FwTemplates []string `json:"fw_templates"`
		Tags        []string `json:"tags"`
	}

	var fwTemplates = make([]string, len(port.FirewallTemplates))
	for i, fwTemplate := range port.FirewallTemplates {
		fwTemplates[i] = fwTemplate.ID
	}
	args := &TempPortCreate{
		Vm:          v.ID,
		Network:     port.Network.ID,
		IpAddress:   port.IpAddress,
		FwTemplates: fwTemplates,
		Tags:        convertTagsToNames(port.Tags),
	}

	if exsist {
		path, _ = url.JoinPath("v1/port", port.ID)
		method = "PUT"
	}

	if err = v.manager.Request(method, path, args, &port); err != nil {
		log.Printf("[REQUEST-ERROR]: connect-port was failed: %s", err)
	} else {
		port.manager = v.manager
	}

	return
}

func (v *Vm) DisconnectPort(port *Port) (err error) {
	path := fmt.Sprintf("v1/port/%s/disconnect", port.ID)

	if err = v.manager.Request("PATCH", path, nil, nil); err != nil {
		log.Printf("[REQUEST-ERROR]: disconnect-port was failed: %s", err)
	} else {
		for i, vmPorts := range v.Ports {
			if vmPorts == port {
				v.Ports = append(v.Ports[:i], v.Ports[i+1:]...)
				break
			}
		}
	}

	return
}

func (v *Vm) PowerOn() error {
	return v.updateState("power_on")
}

func (v *Vm) PowerOff() error {
	return v.updateState("power_off")
}

func (v *Vm) Reboot() error {
	return v.updateState("reboot")
}

func (v *Vm) Reload() (err error) {
	path, _ := url.JoinPath("v1/vm", v.ID)
	m := v.manager

	if err = m.Get(path, Defaults(), &v); err != nil {
		log.Printf("[REQUEST-ERROR] get-vm was failed: %s", err)
	} else {
		v.manager = m
		for x := range v.Ports {
			v.Ports[x].manager = m
		}
		for x := range v.Disks {
			v.Disks[x].manager = m
		}
		v.Vdc.manager = m
		if v.Floating != nil {
			v.Floating.manager = m
		}
	}

	return
}

func (v *Vm) Update() (err error) {
	path, _ := url.JoinPath("v1/vm", v.ID)
	affGr := make([]string, 0)

	if v.AffinityGroups != nil && len(v.AffinityGroups) > 0 {
		for _, group := range v.AffinityGroups {
			affGr = append(affGr, group.ID)
		}
	}

	args := &struct {
		AffinityGroups []string `json:"affinity_groups"`
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		Cpu            int      `json:"cpu"`
		Ram            float64  `json:"ram"`
		HotAdd         bool     `json:"hotadd_feature"`
		Floating       *string  `json:"floating"`
		Tags           []string `json:"tags"`
	}{
		AffinityGroups: affGr,
		Name:           v.Name,
		Description:    v.Description,
		Cpu:            v.Cpu,
		Ram:            v.Ram,
		HotAdd:         v.HotAdd,
		Floating:       nil,
		Tags:           convertTagsToNames(v.Tags),
	}

	if v.Floating != nil {
		if v.Floating.ID != "" {
			args.Floating = &v.Floating.ID
		} else {
			args.Floating = v.Floating.IpAddress
		}
	}

	if err = v.manager.Request("PUT", path, args, v); err != nil {
		log.Printf("[REQUEST-ERROR] update-vm was failed: %s", err)
	}

	return
}

func (v *Vm) updateState(state string) (err error) {
	path := fmt.Sprintf("v1/vm/%s/state", v.ID)

	args := &struct {
		State string `json:"state"`
	}{
		State: state,
	}

	if err = v.manager.Request("POST", path, args, v); err != nil {
		log.Printf("[REQUEST-ERROR] update-vm was failed: %s", err)
	}

	return
}

func (v *Vm) Delete() error {
	path, _ := url.JoinPath("v1/vm", v.ID)
	return v.manager.Delete(path, Defaults(), nil)
}

func (v Vm) WaitLock() error {
	path, _ := url.JoinPath("v1/vm", v.ID)
	return loopWaitLock(v.manager, path)
}
