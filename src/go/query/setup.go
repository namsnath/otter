package query

import (
	"log/slog"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/policy"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

func DeleteEverything() {
	result := db.ExecuteQuery(`MATCH (n) DETACH DELETE n`, nil)
	slog.Info(
		"All nodes and relationships deleted",
		slog.Any("duration", result.Summary.ResultAvailableAfter()),
	)
}

func SetupTestState() {
	db.ExecuteQuery(`CREATE INDEX subject_name_index IF NOT EXISTS FOR (s:Subject) ON (s.name)`, nil)
	db.ExecuteQuery(`CREATE INDEX subject_name_type_index IF NOT EXISTS FOR (s:Subject) ON (s.name, s.type)`, nil)
	db.ExecuteQuery(`CREATE INDEX resource_name_index IF NOT EXISTS FOR (r:Resource) ON (r.name)`, nil)
	db.ExecuteQuery(`CREATE INDEX specifier_key_value_index IF NOT EXISTS FOR (s:Specifier) ON (s.key, s.value)`, nil)
	db.ExecuteQuery(`CREATE INDEX policy_id_index IF NOT EXISTS FOR (p:Policy) ON (p.id)`, nil)

	g2 := subject.Subject{Name: "Group2", Type: subject.SubjectTypeGroup}.Create()
	g1 := subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}.CreateAsChildOf(g2)
	p1 := subject.Subject{Name: "Principal1", Type: subject.SubjectTypePrincipal}.CreateAsChildOf(g1)
	p2 := subject.Subject{Name: "Principal2", Type: subject.SubjectTypePrincipal}.CreateAsChildOf(g2)
	p3 := subject.Subject{Name: "Principal3", Type: subject.SubjectTypePrincipal}.Create()

	rRoot := resource.Resource{Name: "_"}.Create()
	r1 := resource.Resource{Name: "Resource1"}.CreateAsChildOf(rRoot)
	r2 := resource.Resource{Name: "Resource2"}.CreateAsChildOf(rRoot)
	r3 := resource.Resource{Name: "Resource3"}.CreateAsChildOf(rRoot)
	resource.Resource{Name: "Resource4"}.CreateAsChildOf(r3)

	rootSpecifier := specifier.NewSpecifier("*", "*").Create()
	roleRoot, _ := specifier.NewSpecifier("Role", "*").CreateAsChildOf(rootSpecifier)
	roleAdmin, _ := specifier.NewSpecifier("Role", "admin").CreateAsChildOf(roleRoot)
	specifier.NewSpecifier("Role", "user").CreateAsChildOf(roleAdmin)
	envRoot, _ := specifier.NewSpecifier("Env", "*").CreateAsChildOf(rootSpecifier)
	envProd, _ := specifier.NewSpecifier("Env", "prod").CreateAsChildOf(envRoot)
	specifier.NewSpecifier("Env", "dev").CreateAsChildOf(envRoot)

	policies := []policy.Policy{
		{
			Subject:    g1,
			Resource:   r1,
			Action:     action.ActionRead,
			Specifiers: specifier.SpecifierGroup{},
		},
		{
			Subject:    p1,
			Resource:   r3,
			Action:     action.ActionRead,
			Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{envProd}},
		},
		{
			Subject:    p2,
			Resource:   r3,
			Action:     action.ActionRead,
			Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleAdmin, envProd}},
		},
		{
			Subject:    g2,
			Resource:   r2,
			Action:     action.ActionRead,
			Specifiers: specifier.SpecifierGroup{},
		},
		{
			Subject:    p2,
			Resource:   r1,
			Action:     action.ActionRead,
			Specifiers: specifier.SpecifierGroup{},
		},
		{
			Subject:    p3,
			Resource:   rRoot,
			Action:     action.ActionRead,
			Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleAdmin}},
		},
	}

	for _, p := range policies {
		_, err := p.Create()
		if err != nil {
			slog.Error("Error creating policy", "error", err)
		}
	}
}
