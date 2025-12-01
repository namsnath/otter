package query_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/query"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

func TestWhoCanQueries(t *testing.T) {
	query.DeleteEverything()
	query.SetupTestState()

	p1 := subject.Subject{Name: "Principal1", Type: subject.SubjectTypePrincipal}
	p2 := subject.Subject{Name: "Principal2", Type: subject.SubjectTypePrincipal}
	p3 := subject.Subject{Name: "Principal3", Type: subject.SubjectTypePrincipal}
	g1 := subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}
	g2 := subject.Subject{Name: "Group2", Type: subject.SubjectTypeGroup}

	r1 := resource.Resource{Name: "Resource1"}
	r2 := resource.Resource{Name: "Resource2"}
	r3 := resource.Resource{Name: "Resource3"}
	// r4 := resource.Resource{Name: "Resource4"}
	rRoot := resource.Resource{Name: "_"}

	adminRole := specifier.NewSpecifier("Role", "admin")
	// devRole := specifier.NewSpecifier("Role", "dev")
	// devGroup := specifier.SpecifierGroup{Specifiers: []specifier.Specifier{devRole}}
	prodEnv := specifier.NewSpecifier("Env", "prod")
	// devEnv := specifier.NewSpecifier("Env", "dev")

	// query.WhoCan(action.ActionRead).On(r1).Query()
	// query.WhoCan(action.ActionRead).On(r2).Query()
	// query.WhoCan(action.ActionRead).On(rRoot).Query()
	// query.WhoCan(action.ActionRead).On(rRoot).With(specifierGroup).Query()

	testCases := []struct {
		name        string
		subjectType subject.SubjectType
		action      action.Action
		resource    resource.Resource
		specifiers  []specifier.Specifier
		expected    []subject.Subject
	}{
		{"Principals READ r1", subject.SubjectTypePrincipal, action.ActionRead, r1, []specifier.Specifier{}, []subject.Subject{p1, p2}},
		{"Groups READ r1", subject.SubjectTypeGroup, action.ActionRead, r1, []specifier.Specifier{}, []subject.Subject{g1}},
		{"Principals READ r2", subject.SubjectTypePrincipal, action.ActionRead, r2, []specifier.Specifier{}, []subject.Subject{p1, p2}},
		{"Groups READ r2", subject.SubjectTypeGroup, action.ActionRead, r2, []specifier.Specifier{}, []subject.Subject{g1, g2}},
		{"Principals READ rRoot", subject.SubjectTypePrincipal, action.ActionRead, rRoot, []specifier.Specifier{}, []subject.Subject{}},
		{"Groups READ rRoot", subject.SubjectTypeGroup, action.ActionRead, rRoot, []specifier.Specifier{}, []subject.Subject{}},
		{"Principals READ rRoot as admin", subject.SubjectTypePrincipal, action.ActionRead, rRoot, []specifier.Specifier{adminRole}, []subject.Subject{p3}},
		{"Groups READ rRoot as admin", subject.SubjectTypeGroup, action.ActionRead, rRoot, []specifier.Specifier{adminRole}, []subject.Subject{}},
		{"Principals READ r3 as admin", subject.SubjectTypePrincipal, action.ActionRead, r3, []specifier.Specifier{adminRole}, []subject.Subject{p3}},
		{"Groups READ r3 as admin", subject.SubjectTypeGroup, action.ActionRead, r3, []specifier.Specifier{adminRole}, []subject.Subject{}},
		{"Principals READ r3 as admin in prod", subject.SubjectTypePrincipal, action.ActionRead, r3, []specifier.Specifier{adminRole, prodEnv}, []subject.Subject{p1, p2, p3}},
		{"Groups READ r3 as admin in prod", subject.SubjectTypeGroup, action.ActionRead, r3, []specifier.Specifier{adminRole, prodEnv}, []subject.Subject{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := query.WhoCan(tc.subjectType).Perform(tc.action).On(tc.resource).With(specifier.SpecifierGroup{Specifiers: tc.specifiers}).Query()
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
				return
			}

			slices.SortFunc(result, func(a, b subject.Subject) int {
				return strings.Compare(a.Name, b.Name)
			})
			slices.SortFunc(tc.expected, func(a, b subject.Subject) int {
				return strings.Compare(a.Name, b.Name)
			})

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("For %s, expected %v, but got %v", tc.name, tc.expected, result)
			}
		})
	}
}
