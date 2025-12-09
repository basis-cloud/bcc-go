package bcc

import (
	"fmt"
	"net/url"
)

type NodePlatform struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
}

type Kubernetes struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Vdc     *Vdc   `json:"vdc"`
	Vms     []*Vm  `json:"vms"`

	Project       *Project            `json:"project"`
	Floating      *Port               `json:"floating"`
	UserPublicKey string              `json:"user_public_key"`
	Template      *KubernetesTemplate `json:"template"`

	NodeCpu            int             `json:"node_cpu"`
	NodeRam            int             `json:"node_ram"`
	NodesCount         int             `json:"nodes_count"`
	NodePlatform       *Platform       `json:"node_platform"`
	NodeDiskSize       int             `json:"node_disk_size"`
	NodeStorageProfile *StorageProfile `json:"node_storage_profile"`

	Locked bool   `json:"locked"`
	JobId  string `json:"job_id"`
	Tags   []Tag  `json:"tags"`
}

type KubernetesDashBoardUrl struct {
	DashBoardUrl *string `json:"url"`
}

func NewKubernetes(name string, nodeCpu int, nodeRam int, nodesCount int, nodeDiskSize int, floating *string, template *KubernetesTemplate, nodeStorageProfile *StorageProfile, userPublicKey string, nodePlatform *Platform) Kubernetes {
	k := Kubernetes{Name: name, NodeCpu: nodeCpu, NodeDiskSize: nodeDiskSize, NodeRam: nodeRam, NodeStorageProfile: nodeStorageProfile, NodesCount: nodesCount, Template: template, UserPublicKey: userPublicKey, NodePlatform: nodePlatform}
	if floating != nil {
		k.Floating = &Port{IpAddress: floating}
	}
	return k
}

func (m *Manager) ListKubernetes(extraArgs ...Arguments) (ks []*Kubernetes, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/kubernetes"
	err = m.GetItems(path, args, &ks)
	for i := range ks {
		ks[i].manager = m
		for x := range ks[i].Vms {
			ks[i].Vms[x].manager = m
		}
	}
	return
}

func (k *Kubernetes) GetKubernetesConfigUrl() (err error) {
	var config *string
	path := fmt.Sprintf("/v1/kubernetes/%s/config", k.ID)
	err = k.manager.Get(path, Defaults(), &config)
	return
}

func (k *Kubernetes) GetKubernetesDashBoardUrl() (dashboard_url *KubernetesDashBoardUrl, err error) {
	path := fmt.Sprintf("/v1/kubernetes/%s/dashboard", k.ID)
	err = k.manager.Get(path, Defaults(), &dashboard_url)
	return
}

func (m *Manager) GetKubernetes(id string) (k8s *Kubernetes, err error) {
	path, _ := url.JoinPath("/v1/kubernetes", id)
	err = m.Get(path, Defaults(), &k8s)
	if err != nil {
		return
	}
	k8s.Vdc.manager = m
	k8s.manager = m
	return
}

func (v *Vdc) GetKubernetes(extraArgs ...Arguments) (ks []*Kubernetes, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	ks, err = v.manager.ListKubernetes(args)
	return
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
		NodePlatform       *string  `json:"node_platform,omitempty"`
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
		NodePlatform:       nil,
		Tags:               convertTagsToNames(k.Tags),
	}

	if k.Floating != nil {
		args.Floating = k.Floating.IpAddress
	}

	if k.NodePlatform != nil {
		args.NodePlatform = &k.NodePlatform.ID
	}

	if err := v.manager.Request("POST", "/v1/kubernetes", args, &k); err != nil {
		return err
	} else {
		k.manager = v.manager
		for idx := range k.Vms {
			k.Vms[idx].manager = v.manager
		}
	}

	return nil
}

func (k *Kubernetes) Update() error {
	path, _ := url.JoinPath("/v1/kubernetes", k.ID)
	args := &struct {
		Name               string   `json:"name"`
		Floating           *string  `json:"floating"`
		NodesCount         int      `json:"nodes_count"`
		NodesRam           int      `json:"node_ram"`
		NodesCpu           int      `json:"node_cpu"`
		NodeDiskSize       int      `json:"node_disk_size"`
		NodeStorageProfile string   `json:"node_storage_profile"`
		UserPublicKey      string   `json:"user_public_key"`
		Tags               []string `json:"tags"`
	}{
		Name:               k.Name,
		Floating:           nil,
		NodesCount:         k.NodesCount,
		NodesRam:           k.NodeRam,
		NodesCpu:           k.NodeCpu,
		NodeDiskSize:       k.NodeDiskSize,
		NodeStorageProfile: k.NodeStorageProfile.ID,
		UserPublicKey:      k.UserPublicKey,
		Tags:               convertTagsToNames(k.Tags),
	}

	if k.Floating != nil {
		if k.Floating.ID != "" {
			args.Floating = &k.Floating.ID
		} else {
			args.Floating = k.Floating.IpAddress
		}
	}
	err := k.manager.Request("PUT", path, args, k)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kubernetes) Delete() error {
	path, _ := url.JoinPath("v1/kubernetes", k.ID)
	return k.manager.Delete(path, Defaults(), nil)
}

func (k Kubernetes) WaitLock() error {
	path, _ := url.JoinPath("v1/kubernetes", k.ID)
	return loopWaitLock(k.manager, path)
}
