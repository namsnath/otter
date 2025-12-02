package policy

import (
	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

type Policy struct {
	Id         string
	Subject    subject.Subject
	Resource   resource.Resource
	Action     action.Action
	Specifiers specifier.SpecifierGroup
}
