package bcc

import (
	"fmt"
	"log"
	"net/url"
)

type NodePlatform struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
}

type Kubernetes struct {
	manager *Manager
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Vdc     *Vdc     `json:"vdc"`
	Vms     []*TmpVm `json:"vms"`

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

func (m *Manager) ListKubernetes(extraArgs ...Arguments) (k8s []*Kubernetes, err error) {
	path := "v1/kubernetes"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &k8s); err != nil {
		log.Printf("[REQUEST-ERROR] list-kubernetes was failed: %s", err)
	} else {
		for i := range k8s {
			k8s[i].manager = m
			for x := range k8s[i].Vms {
				k8s[i].Vms[x].manager = m
			}
		}
	}

	return
}

func (k *Kubernetes) GetKubernetesConfigUrl() (err error) {
	path := fmt.Sprintf("/v1/kubernetes/%s/config", k.ID)
	var config *string

	if err = k.manager.Get(path, Defaults(), &config); err != nil {
		log.Printf("[REQUEST-ERROR] get-kubernetes-config was failed: %s", err)
	}

	return
}

func (k *Kubernetes) GetKubernetesDashBoardUrl() (dashboardUrl *KubernetesDashBoardUrl, err error) {
	path := fmt.Sprintf("/v1/kubernetes/%s/dashboard", k.ID)

	if err = k.manager.Get(path, Defaults(), &dashboardUrl); err != nil {
		log.Printf("[REQUEST-ERROR] get-kubernetes-dashboard was failed: %s", err)
	}

	return
}

func (m *Manager) GetKubernetes(id string) (k8s *Kubernetes, err error) {
	path, _ := url.JoinPath("/v1/kubernetes", id)

	if err = m.Get(path, Defaults(), &k8s); err != nil {
		log.Printf("[REQUEST-ERROR] get-kubernetes was failed: %s", err)
	} else {
		k8s.Vdc.manager = m
		k8s.manager = m
	}

	return
}

func (v *Vdc) GetKubernetes(extraArgs ...Arguments) (k8s []*Kubernetes, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	k8s, err = v.manager.ListKubernetes(args)
	return
}

func (v *Vdc) CreateKubernetes(k8s *Kubernetes) (err error) {
	path := "/v1/kubernetes"
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
		Name:               k8s.Name,
		NodeCpu:            k8s.NodeCpu,
		NodeRam:            k8s.NodeRam,
		NodeDiskSize:       k8s.NodeDiskSize,
		NodesCount:         k8s.NodesCount,
		NodeStorageProfile: &k8s.NodeStorageProfile.ID,
		Vdc:                &v.ID,
		Template:           &k8s.Template.ID,
		UserPublicKey:      k8s.UserPublicKey,
		Floating:           nil,
		NodePlatform:       nil,
		Tags:               convertTagsToNames(k8s.Tags),
	}

	if k8s.Floating != nil {
		args.Floating = k8s.Floating.IpAddress
	}

	if k8s.NodePlatform != nil {
		args.NodePlatform = &k8s.NodePlatform.ID
	}

	if err = v.manager.Request("POST", path, args, &k8s); err != nil {
		log.Printf("[REQUEST-ERROR] create-kubernetes was failed: %s", err)
	} else {
		k8s.manager = v.manager
		for idx := range k8s.Vms {
			k8s.Vms[idx].manager = v.manager
		}
	}

	return
}

func (k *Kubernetes) Update() (err error) {
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

	if err = k.manager.Request("PUT", path, args, k); err != nil {
		log.Printf("[REQUEST-ERROR] update-kubernetes was failed: %s", err)
	}

	return
}

func (k *Kubernetes) Delete() error {
	path, _ := url.JoinPath("v1/kubernetes", k.ID)
	return k.manager.Delete(path, Defaults(), nil)
}

func (k Kubernetes) WaitLock() error {
	path, _ := url.JoinPath("v1/kubernetes", k.ID)
	return loopWaitLock(k.manager, path)
}
