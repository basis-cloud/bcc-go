package bcc

import (
	"fmt"
	"log"
	"net/url"
)

type FirewallTemplate struct {
	manager     *Manager
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RulesCount  int    `json:"rules_count"`
	Locked      bool   `json:"locked"`
	Tags        []Tag  `json:"tags"`
	Vdc         *Vdc   `json:"vdc,omitempty"`
}

func (m *Manager) GetFirewallTemplate(id string) (firewallTemplate *FirewallTemplate, err error) {
	path, _ := url.JoinPath("v1/firewall/", id)

	if err = m.Get(path, Defaults(), &firewallTemplate); err != nil {
		log.Printf("[REQUEST-ERROR] get-FirewallTemplate with id='%s' was failed: %s", id, err)
	} else {
		firewallTemplate.manager = m
	}

	return
}

func (v *Vdc) GetFirewallTemplates(extraArgs ...Arguments) (firewallTemplate []*FirewallTemplate, err error) {
	path := "v1/firewall"
	args := Arguments{"vdc": v.ID}
	args.merge(extraArgs)

	if err = v.manager.GetItems(path, args, &firewallTemplate); err != nil {
		log.Printf("[REQUEST-ERROR] get-FirewallTemplates failed: %s", err)
	} else {
		for i, _ := range firewallTemplate {
			firewallTemplate[i].manager = v.manager
		}
	}

	return
}

func NewFirewallTemplate(name string) (firewallTemplate FirewallTemplate) {
	d := FirewallTemplate{Name: name}
	return d
}

func (f *FirewallTemplate) Update(firewallRule *FirewallRule) (err error) {
	path := fmt.Sprintf("v1/firewall/%s/rule", f.ID)

	if err = f.manager.Request("POST", path, firewallRule, &firewallRule); err != nil {
		log.Printf("[REQUEST-ERROR] update-FirewallTemplate failed: %s", err)
	} else {
		firewallRule.manager = f.manager
	}

	return
}

func (f *FirewallTemplate) Delete() error {
	path, _ := url.JoinPath("v1/firewall", f.ID)
	return f.manager.Delete(path, Defaults(), nil)
}

func (f *FirewallTemplate) Rename(name string) error {
	f.Name = name
	return f.UpdateFirewallTemplate()
}

func (f *FirewallTemplate) UpdateFirewallTemplate() (err error) {
	path, _ := url.JoinPath("v1/firewall", f.ID)
	args := &struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Tags        []string `json:"tags"`
	}{
		Name:        f.Name,
		Description: f.Description,
		Tags:        convertTagsToNames(f.Tags),
	}

	if err = f.manager.Request("PUT", path, args, &f); err != nil {
		log.Printf("[REQUEST-ERROR] update-FirewallTemplate failed: %s", err)
	}

	return
}

func (v *Vdc) CreateFirewallTemplate(firewallTemplate *FirewallTemplate) (err error) {
	path := "v1/firewall"
	args := &struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Vdc         *string  `json:"vdc,omitempty"`
		Tags        []string `json:"tags"`
	}{
		Name:        firewallTemplate.Name,
		Description: firewallTemplate.Description,
		Vdc:         &v.ID,
		Tags:        convertTagsToNames(firewallTemplate.Tags),
	}

	if err = v.manager.Request("POST", path, args, &firewallTemplate); err != nil {
		log.Printf("[REQUEST-ERROR] create-FirewallTemplate failed: %s", err)
	} else {
		firewallTemplate.manager = v.manager
	}

	return
}

func (f FirewallTemplate) WaitLock() error {
	path, _ := url.JoinPath("v1/firewall", f.ID)
	return loopWaitLock(f.manager, path)
}
