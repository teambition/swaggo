package swaggerv3

// Components ...
type Components struct {
	Schemas map[string]*Schema `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}

// Schema Object allows the definition of input and output data types.
type Schema struct {
	Title                string                `json:"title,omitempty" yaml:"title,omitempty"`
	Ref                  string                `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Format               string                `json:"format,omitempty" yaml:"format,omitempty"`
	Description          string                `json:"description,omitempty" yaml:"description,omitempty"`
	Default              interface{}           `json:"default,omitempty" yaml:"default,omitempty"`
	Required             []string              `json:"required,omitempty" yaml:"required,omitempty"`
	Type                 string                `json:"type,omitempty" yaml:"type,omitempty"`
	Items                *Schema               `json:"items,omitempty" yaml:"items,omitempty"`
	AllOf                []*Schema             `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	Properties           map[string]*Propertie `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *Propertie            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Deprecated           []string              `json:"x-deprecated-fields,omitempty" yaml:"x-deprecated-fields,omitempty"`
}

// Propertie are taken from the JSON Schema definition but their definitions were adjusted to the Swagger Specification
type Propertie struct {
	Ref                  string                `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Title                string                `json:"title,omitempty" yaml:"title,omitempty"`
	Default              interface{}           `json:"default,omitempty" yaml:"default,omitempty"`
	Type                 string                `json:"type,omitempty" yaml:"type,omitempty"`
	Example              string                `json:"example,omitempty" yaml:"example,omitempty"`
	Required             []string              `json:"required,omitempty" yaml:"required,omitempty"`
	ReadOnly             bool                  `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	Properties           map[string]*Propertie `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items                *Propertie            `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalProperties *Propertie            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Description          string                `json:"description,omitempty" yaml:"description,omitempty"`
	Deprecated           []string              `json:"x-deprecated-fields,omitempty" yaml:"x-deprecated-fields,omitempty"`
}
