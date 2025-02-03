package bcc

import "fmt"

type TemplateField struct {
	manager     *Manager
	ID          string `json:"id"`
	Name        string `json:"name"`
	Default     string `json:"default"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Editable    bool   `json:"editable"`
	Position    int    `json:"position"`
	SystemAlias string `json:"system_alias"`
}

func (t *Template) GetFields() (fields []*TemplateField, err error) {
	path := fmt.Sprintf("v1/template/%s/field", t.ID)

	err = t.manager.Get(path, Defaults(), &fields)
	for i := range fields {
		fields[i].manager = t.manager
	}
	return
}
