package bcc

import (
	"fmt"
	"log"
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
	ID                 string        `json:"id"`
	Port               int           `json:"port"`
	CookieName         *string       `json:"cookie_name,omitempty"`
	Connlimit          int           `json:"connlimit"`
	Members            []*PoolMember `json:"members"`
	Method             string        `json:"method,omitempty"`
	Protocol           string        `json:"protocol,omitempty"`
	SessionPersistence string        `json:"session_persistence,omitempty"`

	manager *Manager
	Locked  bool `json:"locked"`
}

type PoolMember struct {
	ID     string `json:"id"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
	Vm     *TmpVm `json:"vm"`
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

func NewLoadBalancerPool(
	lb LoadBalancer, port int, connlimit int, members []*PoolMember,
	method string, protocol string, sessionPersistence string, cookieName interface{}) LoadBalancerPool {

	lbPool := LoadBalancerPool{
		manager:            lb.manager,
		Port:               port,
		Connlimit:          connlimit,
		Members:            members,
		Method:             method,
		Protocol:           protocol,
		SessionPersistence: sessionPersistence,
	}
	if cookieName != nil {
		cookieName := cookieName.(string)
		lbPool.CookieName = &cookieName
	}

	return lbPool
}

func NewLoadBalancerPoolMember(port int, weight int, vm *TmpVm) PoolMember {
	member := PoolMember{
		Weight: weight,
		Vm:     vm,
		Port:   port,
	}
	return member
}

func (m *Manager) GetLoadBalancers(extraArgs ...Arguments) (lbaasList []*LoadBalancer, err error) {
	path := "v1/lbaas"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &lbaasList); err != nil {
		log.Printf("[REQUEST-ERROR]: get-lbaas was failed: %s", err)
	} else {
		for i := range lbaasList {
			lbaasList[i].manager = m
			lbaasList[i].Port.manager = m
			lbaasList[i].Vdc.manager = m
			if lbaasList[i].Floating != nil {
				lbaasList[i].Floating.manager = m
			}
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

	if err = m.Get(path, Defaults(), &lbaas); err != nil {
		log.Printf("[REQUEST-ERROR]: get-lbaas was failed: %s", err)
	} else {
		lbaas.manager = m
		lbaas.Port.manager = m
		lbaas.Vdc.manager = m
		if lbaas.Floating != nil {
			lbaas.Floating.manager = m
		}
	}

	return
}

func (lb *LoadBalancer) Create() (err error) {
	path := "v1/lbaas"
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
	if err = lb.manager.Request("POST", path, lbCreate, &lb); err != nil {
		log.Printf("[REQUEST-ERROR] lbaas.create was failed: %s", err)
	}

	return
}

func (v Vdc) CreateLoadBalancer(lb *LoadBalancer) (err error) {
	path := "v1/lbaas"
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
		lbCreate.Floating = &lb.Floating.ID
	}

	if err = lb.manager.Request("POST", path, lbCreate, &lb); err != nil {
		log.Printf("[REQUEST-ERROR] create-lbaas was failed: %s", err)
	} else {
		lb.manager = v.manager
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
	if err = lb.manager.Request("PUT", path, args, lb); err != nil {
		log.Printf("[REQUEST-ERROR] update-lbaas was failed: %s", err)
	} else {
		err = lb.WaitLock()
	}

	return
}

func (lb *LoadBalancer) Delete() error {
	path, _ := url.JoinPath("v1/lbaas", lb.ID)
	return lb.manager.Delete(path, Defaults(), nil)

}

func (lb *LoadBalancer) GetPools(extraArgs ...Arguments) (pools []*LoadBalancerPool, err error) {
	path := fmt.Sprintf("v1/lbaas/%s/pool", lb.ID)
	args := Defaults()
	args.merge(extraArgs)

	if err = lb.manager.GetSubItems(path, args, &pools); err != nil {
		log.Printf("[REQUEST-ERROR] get-lbaas-pools was failed: %s", err)
	}

	return
}

func (lb *LoadBalancer) GetLoadBalancerPool(id string) (lbaas_pool LoadBalancerPool, err error) {
	path := fmt.Sprintf("v1/lbaas/%s/pool/%s", lb.ID, id)

	if err = lb.manager.Get(path, Defaults(), &lbaas_pool); err != nil {
		log.Printf("[REQUEST-ERROR] get-lbaas-pool was failed: %s", err)
	} else {
		lbaas_pool.manager = lb.manager
	}

	return
}

func (lb *LoadBalancer) CreatePool(pool *LoadBalancerPool) (err error) {
	path := fmt.Sprintf("v1/lbaas/%s/pool", lb.ID)
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
		CookieName         *string       `json:"cookie_name,omitempty"`
		Method             string        `json:"method,omitempty"`
		Protocol           string        `json:"protocol,omitempty"`
		SessionPersistence string        `json:"session_persistence,omitempty"`
	}{
		Port:               pool.Port,
		Connlimit:          pool.Connlimit,
		Members:            members,
		Method:             pool.Method,
		Protocol:           pool.Protocol,
		SessionPersistence: pool.SessionPersistence,
		CookieName:         pool.CookieName,
	}

	if err = lb.manager.Request("POST", path, args, &pool); err != nil {
		log.Printf("[REQUEST-ERROR] create-lbaas-pool was failed: %s", err)
	}

	return
}

func (lb *LoadBalancer) UpdatePool(pool *LoadBalancerPool) (err error) {
	path := fmt.Sprintf("v1/lbaas/%s/pool/%s", lb.ID, pool.ID)

	type poolMember struct {
		Port   int    `json:"port"`
		Weight int    `json:"weight"`
		Vm     string `json:"vm"`
	}
	type createPool struct {
		Port               int           `json:"port"`
		Connlimit          int           `json:"connlimit"`
		Members            []*poolMember `json:"members"`
		Method             string        `json:"method,omitempty"`
		Protocol           string        `json:"protocol,omitempty"`
		CookieName         *string       `json:"cookie_name,omitempty"`
		SessionPersistence string        `json:"session_persistence,omitempty"`
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
		CookieName:         pool.CookieName,
		SessionPersistence: pool.SessionPersistence,
	}

	if err := lb.manager.Request("PUT", path, lbCreatePool, &pool); err != nil {
		log.Printf("[REQUEST-ERROR] update-lbaas-pool was failed: %s", err)
	}

	return
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

func (lb *LoadBalancer) DeletePool(id string) error {
	path := fmt.Sprintf("v1/lbaas/%s/pool/%s", lb.ID, id)
	if err := lb.manager.Delete(path, Defaults(), nil); err != nil {
		return err
	}

	return nil
}

func (lb LoadBalancer) WaitLock() error {
	path, _ := url.JoinPath("v1/lbaas", lb.ID)
	return loopWaitLock(lb.manager, path)
}
