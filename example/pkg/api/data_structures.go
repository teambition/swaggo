package api

import (
	"time"

	"github.com/teambition/swaggo/example/pkg/api/subpackage"
)

type TypeString string

type TypeInterface interface {
	Hello()
}

type SimpleStructure struct {
	Id    float32                    `json:"id" swaggo:"true,dfsdfdsf,19"`
	Name  string                     `json:"name" swaggo:"true,,xus"`
	Age   int                        `json:"age" swaggo:"true,the user age,18"`
	CTime time.Time                  `json:"ctime" swaggo:"true,create time"`
	Sub   subpackage.SimpleStructure `json:"sub" swaggo:"true"`
	I     TypeInterface              `json:"i" swaggo:"true"`
	Map   map[string]string          `json:"map", swaggo:",map type"`
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
