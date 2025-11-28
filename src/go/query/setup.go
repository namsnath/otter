package query

import (
	"log/slog"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/policy"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

func DeleteEverything() {
	result := db.ExecuteQuery(`MATCH (n) DETACH DELETE n`, nil)
	slog.Info(
		"All nodes and relationships deleted",
		slog.Any("resultAvailableAfter", result.Summary.ResultAvailableAfter()),
	)
}

func SetupTestState() {
	// db.ExecuteQuery(ctx, driver, `CREATE CONSTRAINT resource_name_constraint IF NOT EXISTS FOR (r:Resource) REQUIRE r.name IS UNIQUE`, nil)
	db.ExecuteQuery(`CREATE INDEX subject_name_index IF NOT EXISTS FOR (s:Subject) ON (s.name)`, nil)
	db.ExecuteQuery(`CREATE INDEX resource_name_index IF NOT EXISTS FOR (r:Resource) ON (r.name)`, nil)

	g2 := subject.Subject{Name: "Group2", Type: subject.SubjectTypeGroup}.Create()
	g1 := subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}.CreateAsChildOf(g2)
	subject.Subject{Name: "Principal1", Type: subject.SubjectTypePrincipal}.CreateAsChildOf(g1)
	p2 := subject.Subject{Name: "Principal2", Type: subject.SubjectTypePrincipal}.CreateAsChildOf(g2)
	p3 := subject.Subject{Name: "Principal3", Type: subject.SubjectTypePrincipal}.Create()

	rRoot := resource.Resource{Name: "_"}.Create()
	r1 := resource.Resource{Name: "Resource1"}.CreateAsChildOf(rRoot)
	r2 := resource.Resource{Name: "Resource2"}.CreateAsChildOf(rRoot)

	policy.Policy{
		Subject:    g1,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{},
	}.Create()
	policy.Policy{
		Subject:    g2,
		Resource:   r2,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{},
	}.Create()
	policy.Policy{
		Subject:    p2,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{},
	}.Create()
	policy.Policy{
		Subject:    p3,
		Resource:   rRoot,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []*specifier.Specifier{specifier.NewSpecifier("Role", "admin")}},
	}.Create()
}
