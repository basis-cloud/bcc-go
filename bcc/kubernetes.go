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
	Locked  bool   `json:"locked"`
	Vdc     *Vdc   `json:"vdc"`

	Vms     []*Vm `json:"vms"`
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`

	Floating     *Port     `json:"floating"`
	JobId        string    `json:"job_id"`
	NodeCpu      int       `json:"node_cpu"`
	NodeDiskSize int       `json:"node_disk_size"`
	NodePlatform *Platform `json:"node_platform"`

	NodeRam            int                 `json:"node_ram"`
	NodeStorageProfile *StorageProfile     `json:"node_storage_profile"`
	NodesCount         int                 `json:"nodes_count"`
	Template           *KubernetesTemplate `json:"template"`
	UserPublicKey      string              `json:"user_public_key"`
	Tags               []Tag               `json:"tags"`
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

func (v *Vdc) GetKubernetes(extraArgs ...Arguments) (ks []*Kubernetes, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	ks, err = v.manager.ListKubernetes(args)
	return
}

func (k *Kubernetes) GetKubernetesDashBoardUrl() (dashboard_url *KubernetesDashBoardUrl, err error) {
	path := fmt.Sprintf("/v1/kubernetes/%s/dashboard", k.ID)
	err = k.manager.Get(path, Defaults(), &dashboard_url)
	return
}

func (k *Kubernetes) GetKubernetesConfigUrl() (err error) {
	var config *string
	path := fmt.Sprintf("/v1/kubernetes/%s/config", k.ID)
	err = k.manager.Get(path, Defaults(), &config)
	return
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

func (k Kubernetes) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/kubernetes", k.ID)
	return loopWaitLock(k.manager, path)
}
