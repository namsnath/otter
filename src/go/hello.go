package main

import (
	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/query"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

func main() {
	instance := db.GetInstance()
	query.DeleteEverything()
	query.SetupTestState()

	p1 := subject.Subject{Name: "Principal1", Type: subject.SubjectTypePrincipal}
	p2 := subject.Subject{Name: "Principal2", Type: subject.SubjectTypePrincipal}
	p3 := subject.Subject{Name: "Principal3", Type: subject.SubjectTypePrincipal}
	// g1 := subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}
	g2 := subject.Subject{Name: "Group2", Type: subject.SubjectTypeGroup}

	r1 := resource.Resource{Name: "Resource1"}
	r2 := resource.Resource{Name: "Resource2"}
	rRoot := resource.Resource{Name: "_"}

	adminRole := specifier.NewSpecifier("Role", "admin")
	specifierGroup := specifier.SpecifierGroup{Specifiers: []*specifier.Specifier{adminRole}}

	query.WhoCan(action.ActionRead).On(r1).Execute()
	query.WhoCan(action.ActionRead).On(r2).Execute()
	query.WhoCan(action.ActionRead).On(rRoot).Execute()
	query.WhoCan(action.ActionRead).On(rRoot).With(specifierGroup).Execute()

	query.WhatCan(p1).Perform(action.ActionRead).Execute()
	query.WhatCan(g2).Perform(action.ActionRead).Execute()
	query.WhatCan(p2).Perform(action.ActionRead).Execute()
	query.WhatCan(p3).Perform(action.ActionRead).Execute()

	instance.Close()
}
