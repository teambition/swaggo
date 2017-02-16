package api

type SimpleStructure struct {
	Id   int
	Name string
}

type SimpleStructureWithAnnotations struct {
	Id   int    `json:"id"`
	Name string `json:"required,omitempty"`
}

type StructureWithSlice struct {
	Id   int
	Name []byte
}

// hello
type StructureWithEmbededStructure struct {
	StructureWithSlice
}
type StructureWithEmbededPointer struct {
	*StructureWithSlice
}

type APIError struct {
	ErrorCode    int
	ErrorMessage string
}
