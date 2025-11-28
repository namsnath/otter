package main

import (
	"testing"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/query"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

func TestCanQueries(t *testing.T) {
	// instance := db.GetInstance()
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

	testCases := []struct {
		name       string // Optional: for better output with t.Run
		subject    subject.Subject
		action     action.Action
		resource   resource.Resource
		specifiers specifier.SpecifierGroup
		expected   bool
	}{
		{"direct: p1 READ r1", p1, action.ActionRead, r1, specifier.SpecifierGroup{}, true},
		{"direct: p1 READ rRoot", p1, action.ActionRead, rRoot, specifier.SpecifierGroup{}, false},
		{"indirect: p1 READ r2", p1, action.ActionRead, r2, specifier.SpecifierGroup{}, true},
		{"direct: p2 READ r1", p2, action.ActionRead, r1, specifier.SpecifierGroup{}, true},
		{"direct: g2 READ r2", g2, action.ActionRead, r2, specifier.SpecifierGroup{}, true},
		{"direct: g2 READ r2 as admin", g2, action.ActionRead, r2, specifierGroup, false},
		{"indirect: p2 READ r2", p2, action.ActionRead, r2, specifier.SpecifierGroup{}, true},
		{"direct: p3 READ rRoot", p3, action.ActionRead, rRoot, specifier.SpecifierGroup{}, true},
		{"direct: p3 READ rRoot as admin", p3, action.ActionRead, rRoot, specifierGroup, true},
		{"indirect: p3 READ r1 as admin", p3, action.ActionRead, r1, specifierGroup, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := query.Can(tc.subject).Perform(tc.action).On(tc.resource).With(tc.specifiers).Query()
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
				return
			}
			if result != tc.expected {
				t.Errorf("For %s, expected %v, but got %v", tc.name, tc.expected, result)
			}
		})
	}
}
