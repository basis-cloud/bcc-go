package rustack

import (
	"fmt"
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

	path := fmt.Sprintf("v1/firewall/%s/rule", f.ID)
	err = f.manager.Request("POST", path, args, &firewallRule)
	if err != nil {
		return err
	}
	firewallRule.manager = f.manager
	firewallRule.TemplateId = f.ID
	return
}

func (f *FirewallTemplate) GetRuleById(firewallRuleId string) (firewallRule *FirewallRule, err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.ID, firewallRuleId)
	err = f.manager.Get(path, Defaults(), &firewallRule)
	if err != nil {
		return
	}
	firewallRule.manager = f.manager
	firewallRule.TemplateId = f.ID
	return
}

func (m *Manager) GetFirewallRules(id string) (firewallRules []*FirewallRule, err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule", id)
	err = m.Get(path, Defaults(), &firewallRules)
	if err != nil {
		return
	}
	return
}

func (f *FirewallRule) Update() (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.TemplateId, f.ID)
	return f.manager.Request("PUT", path, f, &f)
}

func (f *FirewallRule) Delete() (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.TemplateId, f.ID)
	return f.manager.Delete(path, Defaults(), nil)
}

func (f FirewallRule) WaitLock() (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule/%s", f.TemplateId, f.ID)
	return loopWaitLock(f.manager, path)
}
