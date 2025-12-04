package query_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/query"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

func TestWhatCanQueries(t *testing.T) {
	ctx, container := db.TestContainer()
	// Ensure the container is terminated after the test finishes
	defer func() {
		container.Terminate(ctx)
	}()

	query.DeleteEverything()
	query.SetupTestState()

	p1 := subject.Subject{Name: "Principal1", Type: subject.SubjectTypePrincipal}
	p2 := subject.Subject{Name: "Principal2", Type: subject.SubjectTypePrincipal}
	p3 := subject.Subject{Name: "Principal3", Type: subject.SubjectTypePrincipal}
	// g1 := subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}
	g2 := subject.Subject{Name: "Group2", Type: subject.SubjectTypeGroup}

	r1 := resource.Resource{Name: "Resource1"}
	r2 := resource.Resource{Name: "Resource2"}
	r3 := resource.Resource{Name: "Resource3"}
	r4 := resource.Resource{Name: "Resource4"}
	rRoot := resource.Resource{Name: "_"}

	adminRole := specifier.NewSpecifier("Role", "admin")
	adminGroup := specifier.SpecifierGroup{Specifiers: []specifier.Specifier{adminRole}}
	devRole := specifier.NewSpecifier("Role", "dev")
	devGroup := specifier.SpecifierGroup{Specifiers: []specifier.Specifier{devRole}}
	prodEnv := specifier.NewSpecifier("Env", "prod")
	devEnv := specifier.NewSpecifier("Env", "dev")

	testCases := []struct {
		name       string
		subject    subject.Subject
		action     action.Action
		resource   resource.Resource
		specifiers specifier.SpecifierGroup
		expected   []resource.Resource
	}{
		{"p1 READ", p1, action.ActionRead, resource.Resource{}, specifier.SpecifierGroup{}, []resource.Resource{r1, r2}},
		{"g2 READ", g2, action.ActionRead, resource.Resource{}, specifier.SpecifierGroup{}, []resource.Resource{r2}},
		{"p2 READ", p2, action.ActionRead, resource.Resource{}, specifier.SpecifierGroup{}, []resource.Resource{r1, r2}},
		{"p3 READ", p3, action.ActionRead, resource.Resource{}, specifier.SpecifierGroup{}, []resource.Resource{}},
		{"p3 READ as dev", p3, action.ActionRead, resource.Resource{}, devGroup, []resource.Resource{}},
		{"p3 READ in prod", p3, action.ActionRead, resource.Resource{}, specifier.SpecifierGroup{Specifiers: []specifier.Specifier{prodEnv}}, []resource.Resource{}},
		{"p3 READ in prod as admin", p3, action.ActionRead, resource.Resource{}, specifier.SpecifierGroup{Specifiers: []specifier.Specifier{prodEnv, adminRole}}, []resource.Resource{rRoot, r1, r2, r3, r4}},
		{"p3 READ in dev as admin", p3, action.ActionRead, resource.Resource{}, specifier.SpecifierGroup{Specifiers: []specifier.Specifier{devEnv, adminRole}}, []resource.Resource{rRoot, r1, r2, r3, r4}},
		{"p3 READ as admin", p3, action.ActionRead, resource.Resource{}, adminGroup, []resource.Resource{rRoot, r1, r2, r3, r4}},
		{"p3 READ as admin UNDER r3", p3, action.ActionRead, r3, adminGroup, []resource.Resource{r3, r4}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := query.WhatCan(tc.subject).Perform(tc.action).Under(tc.resource).With(tc.specifiers).Query()
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
				return
			}

			slices.SortFunc(result, func(a, b resource.Resource) int {
				return strings.Compare(a.Name, b.Name)
			})
			slices.SortFunc(tc.expected, func(a, b resource.Resource) int {
				return strings.Compare(a.Name, b.Name)
			})

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("For %s, expected %v, but got %v", tc.name, tc.expected, result)
			}
		})
	}
}
