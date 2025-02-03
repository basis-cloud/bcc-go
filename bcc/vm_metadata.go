package bcc

type VmMetadata struct {
	manager *Manager
	ID      string        `json:"id"`
	Field   TemplateField `json:"field"`
	Value   string        `json:"value"`
}

func NewVmMetadata(field TemplateField, value string) VmMetadata {
	m := VmMetadata{Field: field, Value: value}
	return m
}
