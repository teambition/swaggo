package api

type TypeString string

type SimpleStructure struct {
	Id   float32
	Name string
	Age  int
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
