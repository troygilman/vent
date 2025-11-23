package vent

type VentConfigAnnotation struct {
	VentExtensionConfig
}

func (VentConfigAnnotation) Name() string {
	return "VentConfig"
}

type VentSchemaAnnotation struct {
	TableColumns []string
}

func (VentSchemaAnnotation) Name() string {
	return "VentSchema"
}

type VentFieldAnnotation struct {
}

func (VentFieldAnnotation) Name() string {
	return "VentField"
}
