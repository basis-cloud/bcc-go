package bcc

import (
	"fmt"
	"net/url"
)

type LoadBalancer struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Locked  bool   `json:"locked"`
	Vdc     *Vdc   `json:"vdc"`
	JobId   string `json:"job_id"`

	Kubernetes *Kubernetes `json:"kubernetes"`
	Port       *Port       `json:"port"`
	Floating   *Port       `json:"floating"`
	Tags       []Tag       `json:"tags"`
}

type LoadBalancerPool struct {
	manager *Manager
	ID      string `json:"id"`
	Locked  bool   `json:"locked"`

	Port               int           `json:"port"`
	Connlimit          int           `json:"connlimit"`
	Members            []*PoolMember `json:"members"`
	Method             string        `json:"method"`
	Protocol           string        `json:"protocol"`
	SessionPersistence *string       `json:"session_persistence"`
}

type PoolMember struct {
	ID     string `json:"id"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
	Vm     *Vm    `json:"vm"`
}

func NewLoadBalancer(name string, vdc *Vdc, port *Port, floating *Port) LoadBalancer {
	l := LoadBalancer{
		manager: vdc.manager,
		Name:    name,
		Vdc:     vdc,
		Port:    port,
	}
	if floating != nil {
		l.Floating = floating
	}
	return l
}

func (lb *LoadBalancer) Create() (err error) {
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
		Floating   *string     `json:"floating,omitempty"`
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
		if lb.Floating.ID != "" {
			lbCreate.Floating = &lb.Floating.ID
		} else {
			lbCreate.Floating = lb.Floating.IpAddress
		}
	}
	err = lb.manager.Request("POST", "v1/lbaas", lbCreate, &lb)
	return
}

func (m *Manager) GetLoadBalancers(extraArgs ...Arguments) (lbaasList []*LoadBalancer, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/lbaas"
	err = m.GetItems(path, args, &lbaasList)
	if err != nil {
		return
	}
	for i := range lbaasList {
		lbaasList[i].manager = m
		lbaasList[i].Port.manager = m
		lbaasList[i].Vdc.manager = m
		if lbaasList[i].Floating != nil {
			lbaasList[i].Floating.manager = m
		}
	}
	return
}

func (v *Vdc) GetLoadBalancers(extraArgs ...Arguments) (lbaasList []*LoadBalancer, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	lbaasList, err = v.manager.GetLoadBalancers(args)
	return
}

func (m *Manager) GetLoadBalancer(id string) (lbaas *LoadBalancer, err error) {
	path, _ := url.JoinPath("v1/lbaas", id)
	err = m.Get(path, Defaults(), &lbaas)
	if err != nil {
		return
	}
	lbaas.manager = m
	lbaas.Port.manager = m
	lbaas.Vdc.manager = m
	if lbaas.Floating != nil {
		lbaas.Floating.manager = m
	}
	return
}

func (lb *LoadBalancer) Update() (err error) {
	path, _ := url.JoinPath("v1/lbaas", lb.ID)
	args := &struct {
		Name     string  `json:"name"`
		Floating *string `json:"floating"`
		Port
		Tags []string `json:"tags"`
	}{
		Name:     lb.Name,
		Floating: nil,
		Tags:     convertTagsToNames(lb.Tags),
	}

	if lb.Floating != nil {
		if lb.Floating.ID != "" {
			args.Floating = &lb.Floating.ID
		} else {
			args.Floating = lb.Floating.IpAddress
		}
	}
	err = lb.manager.Request("PUT", path, args, lb)
	lb.WaitLock()
	return
}

func (lb *LoadBalancer) Delete() (err error) {
	path, _ := url.JoinPath("v1/lbaas", lb.ID)
	return lb.manager.Delete(path, Defaults(), nil)

}

func NewLoadBalancerPool(lb LoadBalancer, port int, connlimit int, members []*PoolMember, method string, protocol string, session_persistence string) LoadBalancerPool {
	lb_pool := LoadBalancerPool{
		manager:            lb.manager,
		Port:               port,
		Connlimit:          connlimit,
		Members:            members,
		Method:             method,
		Protocol:           protocol,
		SessionPersistence: &session_persistence,
	}

	return lb_pool
}

func NewLoadBalancerPoolMember(port int, weight int, vm *Vm) PoolMember {
	member := PoolMember{
		Weight: weight,
		Vm:     vm,
		Port:   port,
	}
	return member
}

func (lb *LoadBalancer) CreatePool(pool *LoadBalancerPool) (err error) {
	type poolMember struct {
		Port   int    `json:"port"`
		Weight int    `json:"weight"`
		Vm     string `json:"vm"`
	}
	var members []*poolMember
	for _, member := range pool.Members {
		members = append(members, &poolMember{
			Port:   member.Port,
			Weight: member.Weight,
			Vm:     member.Vm.ID,
		})
	}

	args := &struct {
		Port               int           `json:"port"`
		Connlimit          int           `json:"connlimit"`
		Members            []*poolMember `json:"members"`
		Method             string        `json:"method"`
		Protocol           string        `json:"protocol"`
		SessionPersistence *string       `json:"session_persistence"`
	}{
		Port:               pool.Port,
		Connlimit:          pool.Connlimit,
		Members:            members,
		Method:             pool.Method,
		Protocol:           pool.Protocol,
		SessionPersistence: nil,
	}

	if pool.SessionPersistence != nil && *pool.SessionPersistence != "" {
		args.SessionPersistence = pool.SessionPersistence
	}

	path := fmt.Sprintf("v1/lbaas/%s/pool", lb.ID)
	err = lb.manager.Request("POST", path, args, &pool)
	return
}

func (lb *LoadBalancer) UpdatePool(pool *LoadBalancerPool) (err error) {
	type poolMember struct {
		Port   int    `json:"port"`
		Weight int    `json:"weight"`
		Vm     string `json:"vm"`
	}
	type createPool struct {
		Port               int           `json:"port"`
		Connlimit          int           `json:"connlimit"`
		Members            []*poolMember `json:"members"`
		Method             string        `json:"method"`
		Protocol           string        `json:"protocol"`
		SessionPersistence *string       `json:"session_persistence"`
	}

	var members []*poolMember
	for _, member := range pool.Members {
		members = append(members, &poolMember{
			Port:   member.Port,
			Weight: member.Weight,
			Vm:     member.Vm.ID,
		})
	}

	lbCreatePool := createPool{
		Port:               pool.Port,
		Connlimit:          pool.Connlimit,
		Members:            members,
		Method:             pool.Method,
		Protocol:           pool.Protocol,
		SessionPersistence: pool.SessionPersistence,
	}
	path := fmt.Sprintf("v1/lbaas/%s/pool/%s", lb.ID, pool.ID)
	err = lb.manager.Request("PUT", path, lbCreatePool, &pool)
	return
}

func (lb *LoadBalancer) GetLoadBalancerPool(id string) (lbaas_pool LoadBalancerPool, err error) {
	path := fmt.Sprintf("v1/lbaas/%s/pool/%s", lb.ID, id)
	err = lb.manager.Get(path, Defaults(), &lbaas_pool)
	if err != nil {
		return
	}
	lbaas_pool.manager = lb.manager
	return
}

func (lb *LoadBalancer) GetPools() (pools []*LoadBalancerPool, err error) {
	path := fmt.Sprintf("v1/lbaas/%s/pool", lb.ID)
	err = lb.manager.GetSubItems(path, Arguments{}, &pools)
	return pools, err
}

func (lb *LoadBalancer) DeletePools() error {
	pools, err := lb.GetPools()
	if err != nil {
		return err
	}
	for _, pool := range pools {
		err = lb.DeletePool(pool.ID)
		if err != nil {
			return err
		}
		lb.WaitLock()
	}
	return nil
}

func (lb *LoadBalancer) DeletePool(id string) (err error) {
	path := fmt.Sprintf("v1/lbaas/%s/pool/%s", lb.ID, id)
	err = lb.manager.Delete(path, Defaults(), Defaults())
	return
}

func (lb LoadBalancer) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/lbaas", lb.ID)
	return loopWaitLock(lb.manager, path)
}
