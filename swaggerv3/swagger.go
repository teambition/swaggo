package swaggerv3

// New ...
func New() *Swagger {
	return &Swagger{}
}

// Swagger ...
type Swagger struct {
	Openapi    string              `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	Info       Info                `json:"info,omitempty" yaml:"info,omitempty"`
	Tags       []Tag               `json:"tags,omitempty" yaml:"tags,omitempty"`
	Paths      map[string]PathItem `json:"paths,omitempty" yaml:"paths,omitempty"`
	Components Components          `json:"components,omitempty" yaml:"components,omitempty"`
	Servers    []Server            `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// Info ...
type Info struct {
	Version        string  `json:"version,omitempty" yaml:"version,omitempty"`
	Title          string  `json:"title,omitempty" yaml:"title,omitempty"`
	Description    string  `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string  `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
}

// Contact ...
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// Tag ...
type Tag struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Server ...
type Server struct {
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}
