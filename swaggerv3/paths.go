package swaggerv3

// PathItem Describes the operations available on a single path.
type PathItem struct {
	Ref     string     `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Get     *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Put     *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Post    *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Delete  *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options *Operation `json:"options,omitempty" yaml:"options,omitempty"`
	Head    *Operation `json:"head,omitempty" yaml:"head,omitempty"`
	Patch   *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
}

// Operation Describes a single API operation on a path.
type Operation struct {
	Tags        []string             `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string               `json:"summary,omitempty" yaml:"summary,omitempty"`
	Permissions []Permission         `json:"x-permissions,omitempty" yaml:"x-permissions,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Consumes    []string             `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces    []string             `json:"produces,omitempty" yaml:"produces,omitempty"`
	Schemes     []string             `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Parameters  []*Parameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]*Response `json:"responses,omitempty" yaml:"responses,omitempty"`
	Deprecated  bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// Permission ...
type Permission struct {
	Resource string `json:"resource,omitempty" yaml:"resource,omitempty"`
	Action   string `json:"action,omitempty" yaml:"action,omitempty"`
}

// Parameter Describes a single operation parameter.
type Parameter struct {
	In          string           `json:"in,omitempty" yaml:"in,omitempty"`
	Name        string           `json:"name,omitempty" yaml:"name,omitempty"`
	Description string           `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool             `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *ParameterSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// ParameterSchema Object allows the definition of input and output data types.
type ParameterSchema struct {
	Type   string           `json:"type,omitempty" yaml:"type,omitempty"`
	Format string           `json:"format,omitempty" yaml:"format,omitempty"`
	Ref    string           `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Items  *ParameterSchema `json:"items,omitempty" yaml:"items,omitempty"`
}

// RequestBody ...
type RequestBody struct {
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Content     BodyContent `json:"content,omitempty" yaml:"content,omitempty"`
}

// BodyContent ...
type BodyContent struct {
	ApplicationJSON ApplicationJSON `json:"application/json,omitempty" yaml:"application/json,omitempty"`
}

// ApplicationJSON ...
type ApplicationJSON struct {
	Schema *PathSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// PathSchema Object allows the definition of input and output data types.
type PathSchema struct {
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Ref         string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type        string      `json:"type,omitempty" yaml:"type,omitempty"`
	Items       *PathSchema `json:"items,omitempty" yaml:"items,omitempty"`
}

// Response as they are returned from executing this operation.
type Response struct {
	Description string  `json:"description" yaml:"description"`
	Content     Content `json:"content,omitempty" yaml:"content,omitempty"`
	Ref         string  `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

// Content ...
type Content struct {
	ApplicationJSON ApplicationJSON `json:"application/json,omitempty" yaml:"application/json,omitempty"`
}
