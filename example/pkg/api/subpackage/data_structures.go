package subpackage

type SimpleStructure struct {
	Id   int    `json:"id" required:"true" default:"2" desc:"the user id"`
	Name string `json:"name" required:"true" default:"John Smith" desc:"the user name"`
}
