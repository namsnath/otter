package main

import (
	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/query"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/subject"
)

func main() {
	instance := db.GetInstance()
	query.DeleteEverything()
	query.SetupTestState()

	p1 := &subject.Subject{Name: "Principal1", Type: subject.SubjectTypePrincipal}
	p2 := &subject.Subject{Name: "Principal2", Type: subject.SubjectTypePrincipal}
	p3 := &subject.Subject{Name: "Principal3", Type: subject.SubjectTypePrincipal}
	// g1 := &subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}
	g2 := &subject.Subject{Name: "Group2", Type: subject.SubjectTypeGroup}

	r1 := &resource.Resource{Name: "Resource1"}
	r2 := &resource.Resource{Name: "Resource2"}
	rRoot := &resource.Resource{Name: "_"}

	query.SubjectCanDo(p1, action.ActionRead, r1, nil)
	query.SubjectCanDo(g2, action.ActionRead, r2, nil)
	query.SubjectCanDo(p3, action.ActionRead, rRoot, nil)

	query.AllSubjectsThatCanDo(r1, action.ActionRead, nil)
	query.AllSubjectsThatCanDo(r2, action.ActionRead, nil)
	query.AllSubjectsThatCanDo(rRoot, action.ActionRead, nil)

	query.AllResourcesThatSubjectCanDo(p1, action.ActionRead, nil)
	query.AllResourcesThatSubjectCanDo(g2, action.ActionRead, nil)
	query.AllResourcesThatSubjectCanDo(p2, action.ActionRead, nil)
	query.AllResourcesThatSubjectCanDo(p3, action.ActionRead, nil)

	instance.Close()
}
