package bcc

import (
	"log"
	"net/url"
)

type Project struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Client  struct {
		Id string `json:"id"`
	} `json:"client"`
	Locked bool  `json:"locked"`
	Tags   []Tag `json:"tags"`
}

func NewProject(name string) Project {
	b := Project{Name: name}
	return b
}

func (m *Manager) GetProjects(extraArgs ...Arguments) (projects []*Project, err error) {
	path := "v1/project"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &projects); err != nil {
		log.Printf("[REQUEST-ERROR]: get-projects was failed: %s", err)
	} else {
		for i := range projects {
			projects[i].manager = m
		}
	}

	return
}

func (m *Manager) GetProject(id string) (project *Project, err error) {
	path, _ := url.JoinPath("v1/project", id)

	if err = m.Get(path, Defaults(), &project); err != nil {
		log.Printf("[REQUEST-ERROR]: getting project-%s was failed: %s]", id, err)
	} else {
		project.manager = m
	}

	return
}

func (c *Client) CreateProject(project *Project) (err error) {
	path := "v1/project"
	args := &struct {
		Name   string   `json:"name"`
		Client string   `json:"client"`
		Tags   []string `json:"tags"`
	}{
		Name:   project.Name,
		Client: c.ID,
		Tags:   convertTagsToNames(project.Tags),
	}

	if err := c.manager.Request("POST", path, args, &project); err != nil {
		log.Printf("[REQUEST-ERROR]: creating project-%s was failed: %s", project.Name, err)
	} else {
		project.manager = c.manager
	}

	return
}

func (p *Project) Rename(name string) error {
	p.Name = name
	return p.Update()
}

func (p *Project) Update() (err error) {
	path, _ := url.JoinPath("v1/project", p.ID)
	args := &struct {
		Name   string   `json:"name"`
		Client string   `json:"client"`
		Tags   []string `json:"tags"`
	}{
		Name:   p.Name,
		Client: p.Client.Id,
		Tags:   convertTagsToNames(p.Tags),
	}

	if err = p.manager.Request("PUT", path, args, p); err != nil {
		log.Printf("[REQUEST-ERROR]: updating project-%s was failed: %s", p.Name, err)
	}

	return
}

func (p *Project) Delete() error {
	path, _ := url.JoinPath("v1/project", p.ID)
	return p.manager.Delete(path, Defaults(), nil)
}

func (p Project) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/project", p.ID)
	return loopWaitLock(p.manager, path)
}
