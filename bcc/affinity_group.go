package bcc

import "net/url"

type AffinityGroup struct {
	manager     *Manager
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Policy      string      `json:"policy"`
	Vms         []*MetaData `json:"vms,omitempty"`
	Vdc         *Vdc        `json:"vdc"`
	Locked      bool        `json:"locked,omitempty"`
	JobId       string      `json:"job_id,omitempty"`
}

func NewAffinityGroup(name string, description string, policy string, vms []*MetaData) AffinityGroup {
	return AffinityGroup{Name: name, Description: description, Policy: policy, Vms: vms}
}

func (m *Manager) GetAffinityGroups(extraArgs ...Arguments) (affinityGroups []*AffinityGroup, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/affinity_group"
	err = m.GetItems(path, args, &affinityGroups)

	for i := range affinityGroups {
		affinityGroups[i].manager = m
	}

	return
}

func (v *Vdc) GetAffinityGroups(extraArgs ...Arguments) (affinityGroups []*AffinityGroup, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	affinityGroups, err = v.manager.GetAffinityGroups(args)
	return
}

func (m *Manager) GetAffinityGroup(id string) (affinityGroup *AffinityGroup, err error) {
	path, _ := url.JoinPath("v1/affinity_group", id)
	if err = m.Get(path, Defaults(), &affinityGroup); err != nil {
		return
	}

	affinityGroup.manager = m

	return
}

func (v *Vdc) CreateAffinityGroup(affinityGroup *AffinityGroup) error {
	args := &struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Policy      string   `json:"policy"`
		Vms         []string `json:"vms,omitempty"`
		Vdc         string   `json:"vdc"`
	}{
		Name:        affinityGroup.Name,
		Description: affinityGroup.Description,
		Policy:      affinityGroup.Policy,
		Vms:         convertNameToId(affinityGroup.Vms),
		Vdc:         v.ID,
	}

	if err := v.manager.Request("POST", "v1/affinity_group", args, &affinityGroup); err != nil {
		return err
	} else {
		affinityGroup.manager = v.manager
		affinityGroup.Vdc = v
	}

	return nil
}

func (a *AffinityGroup) Reload() error {
	m := a.manager

	path, _ := url.JoinPath("v1/affinity_group", a.ID)
	if err := m.Get(path, Defaults(), &a); err != nil {
		return err
	}

	a.manager = m
	a.Vdc.manager = m

	return nil
}

func (a *AffinityGroup) Update() error {
	path, _ := url.JoinPath("v1/affinity_group", a.ID)

	args := &struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Policy      string   `json:"policy"`
		Vms         []string `json:"vms,omitempty"`
	}{
		Name:        a.Name,
		Description: a.Description,
		Policy:      a.Policy,
		Vms:         convertNameToId(a.Vms),
	}

	if err := a.manager.Request("PUT", path, args, a); err != nil {
		return err
	}

	return nil
}

func (a *AffinityGroup) Delete() error {
	path, _ := url.JoinPath("v1/affinity_group", a.ID)
	return a.manager.Delete(path, Defaults(), nil)
}

func (a *AffinityGroup) WaitLock() error {
	path, _ := url.JoinPath("v1/affinity_group", a.ID)
	return loopWaitLock(a.manager, path)
}
