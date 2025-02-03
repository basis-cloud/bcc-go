package bcc

import (
	"net/url"
	"strconv"
)

type PaasInputDescription struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Value       string                 `json:"value"`
	Required    bool                   `json:"required"`
	Default     interface{}            `json:"default"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type PaasTemplate struct {
	manager     *Manager
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PaasService struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	PaasDeployID    int                    `json:"paas_deploy_id,omitempty"`
	PaasServiceID   int                    `json:"paas_service_id"`
	PaasServiceName string                 `json:"paas_service_name"`
	Status          string                 `json:"status,omitempty"`
	PaasInternalID  string                 `json:"paas_internal_id,omitempty"`
	Inputs          map[string]interface{} `json:"paas_service_inputs"`
	Locked          bool                   `json:"locked"`
}

func (m *Manager) GetPaasTemplates(projectId string) ([]*PaasTemplate, error) {
	var templates []*PaasTemplate
	args := Arguments{"project_id": projectId}
	path := "v1/paas_template"
	if err := m.GetItems(path, args, &templates); err != nil {
		return nil, err
	}
	for i := range templates {
		templates[i].manager = m
	}
	return templates, nil
}

func (m *Manager) GetPaasTemplate(id int, projectId string) (*PaasTemplate, error) {
	var template *PaasTemplate
	args := Arguments{"project_id": projectId}
	path, _ := url.JoinPath("v1/paas_template", strconv.Itoa(id))
	if err := m.Get(path, args, &template); err != nil {
		return nil, err
	}
	template.manager = m
	return template, nil
}

func (p *PaasTemplate) GetPaasTemplateInputs(projectId string) ([]*PaasInputDescription, error) {
	path, _ := url.JoinPath("v1/paas_template", strconv.Itoa(p.ID), "inputs")
	response := struct {
		Inputs []*PaasInputDescription `json:"inputs"`
	}{}
	args := Arguments{"project_id": projectId}
	if err := p.manager.Request("GET", path, args, &response); err != nil {
		return nil, err
	}
	return response.Inputs, nil
}

func (m *Manager) GetPaasServices(args Arguments) ([]*PaasService, error) {
	var services []*PaasService
	path := "v1/paas_service"
	if err := m.GetItems(path, args, &services); err != nil {
		return nil, err
	}
	for i := range services {
		services[i].manager = m
	}
	return services, nil
}

func (m *Manager) GetPaasService(id string) (*PaasService, error) {
	var service *PaasService
	path, _ := url.JoinPath("v1/paas_service", id)
	if err := m.Get(path, Defaults(), &service); err != nil {
		return nil, err
	}
	service.manager = m
	return service, nil
}

func (m *Manager) CreatePaasService(p *PaasService) error {
	args := struct {
		Name          string                 `json:"name"`
		Project       string                 `json:"project"`
		PaasServiceID int                    `json:"paas_service_id"`
		Inputs        map[string]interface{} `json:"paas_service_inputs"`
	}{
		Name:          p.Name,
		Project:       p.Project.ID,
		PaasServiceID: p.PaasServiceID,
		Inputs:        p.Inputs,
	}
	path := "v1/paas_service"
	err := m.Request("POST", path, args, &p)
	if err != nil {
		return err
	}
	p.manager = m
	return nil
}

func (m *Manager) DeletePaasService(id string) error {
	path, _ := url.JoinPath("v1/paas_service", id)
	return m.Delete(path, Defaults(), nil)
}
