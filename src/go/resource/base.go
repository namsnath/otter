package resource

type Resource struct {
	Name string
}

func NewResource(name string) Resource {
	return Resource{Name: name}
}
