package bcc

import (
	"fmt"
	"log"
)

type FirewallRule struct {
	manager         *Manager
	TemplateId      string
	ID              string `json:"id"`
	Name            string `json:"name"`
	DestinationIp   string `json:"destination_ip"`
	Direction       string `json:"direction"`
	DstPortRangeMax *int   `json:"dst_port_range_max"`
	DstPortRangeMin *int   `json:"dst_port_range_min"`
	Protocol        string `json:"protocol"`
	Locked          bool   `json:"locked"`
}

func NewFirewallRule(name string, destinationIp string, direction string, protocol string, dstPortRangeMax int, dstPortRangeMin int) (firewallRule FirewallRule) {
	d := FirewallRule{
		Name:            name,
		DestinationIp:   destinationIp,
		Direction:       direction,
		DstPortRangeMax: &dstPortRangeMax,
		DstPortRangeMin: &dstPortRangeMin,
		Protocol:        protocol,
	}
	return d
}

func (f *FirewallTemplate) CreateFirewallRule(firewallRule *FirewallRule) (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule", f.ID)
	args := &struct {
		manager         *Manager
		ID              string `json:"id"`
		Name            string `json:"name"`
		DestinationIp   string `json:"destination_ip"`
		Direction       string `json:"direction"`
		DstPortRangeMax *int   `json:"dst_port_range_max"`
		DstPortRangeMin *int   `json:"dst_port_range_min"`
		Protocol        string `json:"protocol"`
	}{
		ID:              firewallRule.ID,
		Name:            firewallRule.Name,
		DestinationIp:   firewallRule.DestinationIp,
		Direction:       firewallRule.Direction,
		DstPortRangeMax: nil,
		DstPortRangeMin: nil,
		Protocol:        firewallRule.Protocol,
	}

	if firewallRule.Protocol == "tcp" || firewallRule.Protocol == "udp" {
		args.DstPortRangeMax = firewallRule.DstPortRangeMax
		args.DstPortRangeMin = firewallRule.DstPortRangeMin
	}

	if err = f.manager.Request("POST", path, args, &firewallRule); err != nil {
		log.Printf("[REQUEST-ERROR] create-FirewallRule was failed: %s", err)
	} else {
		firewallRule.manager = f.manager
		firewallRule.TemplateId = f.ID
	}

	return
}

func (f *FirewallTemplate) GetRuleById(firewallRuleId string) (firewallRule *FirewallRule, err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.ID, firewallRuleId)

	if err = f.manager.Get(path, Defaults(), &firewallRule); err != nil {
		log.Printf("[REQUEST-ERROR] get-Firewall rule was failed: %s", err)
	} else {
		firewallRule.manager = f.manager
		firewallRule.TemplateId = f.ID
	}

	return
}

func (m *Manager) GetFirewallRules(id string, extraArgs ...Arguments) (firewallRules []*FirewallRule, err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule", id)
	args := Defaults()
	args.merge(extraArgs)

	if err = m.Get(path, args, &firewallRules); err != nil {
		log.Printf("[REQUEST-ERROR] get-Firewall rules was failed: %s", err)
	}

	return
}

func (f *FirewallRule) Update() (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.TemplateId, f.ID)

	if err = f.manager.Request("PUT", path, f, &f); err != nil {
		log.Printf("[REQUEST-ERROR] update-FirewallRule was failed: %s", err)
	}

	return
}

func (f *FirewallRule) Delete() (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.TemplateId, f.ID)

	if err = f.manager.Delete(path, Defaults(), nil); err != nil {
		log.Printf("[REQUEST-ERROR] delete-FirewallRule was failed: %s", err)
	}

	return
}

func (f FirewallRule) WaitLock() (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.TemplateId, f.ID)
	return loopWaitLock(f.manager, path)
}
