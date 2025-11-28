package policy

import (
	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

type Policy struct {
	Subject    subject.Subject
	Resource   resource.Resource
	Action     action.Action
	Specifiers specifier.SpecifierGroup
}
