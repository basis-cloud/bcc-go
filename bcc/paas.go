package bcc

import (
	"log"
	"net/url"
)

type PaasInputDescription struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Value       string                 `json:"value"`
	Required    bool                   `json:"required"`
	Default     interface{}            `json:"default"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type PaasTemplate struct {
	manager      *Manager
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	DisplayName  string   `json:"display_name"`
	Tenant       string   `json:"tenant"`
	Base64Icon   string   `json:"base64_icon"`
	PlatformTags []string `json:"platform_tags"`
	Platforms    []int    `json:"platforms"`
	Tags         []string `json:"tags"`

	PublishedToShowcase bool `json:"published_to_showcase"`
}

type PaasService struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Vdc     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"vdc"`
	PaasDeployID    int                    `json:"paas_deploy_id,omitempty"`
	PaasServiceID   int                    `json:"paas_service_id"`
	PaasServiceName string                 `json:"paas_service_name"`
	Status          string                 `json:"status,omitempty"`
	PaasInternalID  string                 `json:"paas_internal_id,omitempty"`
	Inputs          map[string]interface{} `json:"paas_service_inputs"`
	Locked          bool                   `json:"locked"`
}

func (m *Manager) CreatePaasLocation(vdcId string) (err error) {
	path := "v1/paas"
	args := struct {
		Vdc string `json:"vdc"`
	}{
		Vdc: vdcId,
	}

	if err = m.Request("POST", path, args, nil); err != nil {
		log.Printf("[REQUEST-ERROR]: creating paas location was failed: %s", err)
	}

	return
}

func (m *Manager) GetPaasTemplates(vdcId string, extraArgs ...Arguments) (templates []*PaasTemplate, err error) {
	path := "v1/paas_template"
	args := Arguments{"vdc_id": vdcId}
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &templates); err != nil {
		log.Printf("[REQUEST-ERROR]: get-paas-templates was failed: %s", err)
	} else {
		for i := range templates {
			templates[i].manager = m
		}
	}

	return
}

func (m *Manager) GetPaasTemplate(id string, vdcId string) (template *PaasTemplate, err error) {
	path, _ := url.JoinPath("v1/paas_template", id)
	args := Arguments{"vdc_id": vdcId}

	if err = m.Get(path, args, &template); err != nil {
		log.Printf("[REQUEST-ERROR]: get-paas-template was failed: %s", err)
	} else {
		template.manager = m
	}

	return
}

func (p *PaasTemplate) GetPaasTemplateInputs(projectId string, extraArgs ...Arguments) ([]*PaasInputDescription, error) {
	path, _ := url.JoinPath("v1/paas_template", p.ID, "inputs")
	response := struct {
		Inputs []*PaasInputDescription `json:"inputs"`
	}{}
	args := Arguments{"project_id": projectId}
	args.merge(extraArgs)

	if err := p.manager.Request("GET", path, args, &response); err != nil {
		log.Printf("[REQUEST-ERROR]: get-paas-template-inputs was failed: %s", err)
	}

	return response.Inputs, nil
}

func (m *Manager) GetPaasServices(args Arguments) (services []*PaasService, err error) {
	path := "v1/paas_service"

	if err = m.GetItems(path, args, &services); err != nil {
		log.Printf("[REQUEST-ERROR]: get-paas-services was failed: %s", err)
	} else {
		for i := range services {
			services[i].manager = m
		}
	}

	return
}

func (m *Manager) GetPaasService(id string) (service *PaasService, err error) {
	path, _ := url.JoinPath("v1/paas_service", id)

	if err := m.Get(path, Defaults(), &service); err != nil {
		log.Printf("[REQUEST-ERROR]: get-paas-service was failed: %s", err)
	} else {
		service.manager = m
	}

	return
}

func (m *Manager) CreatePaasService(p *PaasService) error {
	path := "v1/paas_service"
	args := struct {
		Name          string                 `json:"name"`
		Vdc           string                 `json:"vdc"`
		PaasServiceID int                    `json:"paas_service_id"`
		Inputs        map[string]interface{} `json:"paas_service_inputs"`
	}{
		Name:          p.Name,
		Vdc:           p.Vdc.ID,
		PaasServiceID: p.PaasServiceID,
		Inputs:        p.Inputs,
	}

	if err := m.Request("POST", path, args, &p); err != nil {
		log.Printf("[REQUEST-ERROR]: creating paas service was failed: %s", err)
	} else {
		p.manager = m
	}

	return nil
}

func (p *PaasService) Update() (err error) {
	path, _ := url.JoinPath("v1/paas_service", p.ID)
	args := struct {
		Name   string                 `json:"name"`
		Inputs map[string]interface{} `json:"paas_service_inputs"`
	}{
		Name:   p.Name,
		Inputs: p.Inputs,
	}

	if err = p.manager.Request("PUT", path, args, p); err != nil {
		log.Printf("[REQUEST-ERROR]: updating paas service was failed: %s", err)
	}

	return
}

func (m *Manager) DeletePaasService(id string) error {
	path, _ := url.JoinPath("v1/paas_service", id)
	return m.Delete(path, Defaults(), nil)
}

func (p PaasService) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/paas_service", p.ID)
	return loopWaitLock(p.manager, path)
}
