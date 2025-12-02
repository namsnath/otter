package policy_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/policy"
	"github.com/namsnath/otter/query"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

func TestPolicyGetQueries(t *testing.T) {
	query.DeleteEverything()
	query.SetupIndexes()

	g1 := subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}.Create()
	g2 := subject.Subject{Name: "Group2", Type: subject.SubjectTypeGroup}.Create()
	rRoot := resource.Resource{Name: "_"}.Create()
	r1 := resource.Resource{Name: "Resource1"}.CreateAsChildOf(rRoot)
	rootSpecifier := specifier.NewSpecifier("*", "*").Create()
	envRoot, _ := specifier.NewSpecifier("Env", "*").CreateAsChildOf(rootSpecifier)
	envProd, _ := specifier.NewSpecifier("Env", "prod").CreateAsChildOf(envRoot)
	specifier.NewSpecifier("Env", "dev").CreateAsChildOf(envRoot)

	g1R1ReadPolicy, _ := policy.Policy{
		Subject:    g1,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{},
	}.Create()
	expectedG1R1ReadPolicy := g1R1ReadPolicy
	expectedG1R1ReadPolicy.Specifiers = specifier.SpecifierGroup{Specifiers: []specifier.Specifier{envRoot}}

	g2R1ProdReadPolicy, _ := policy.Policy{
		Subject:    g2,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{envProd}},
	}.Create()
	expectedG2R1ProdReadPolicy := g2R1ProdReadPolicy

	g1R1WritePolicy, _ := policy.Policy{
		Subject:    g1,
		Resource:   r1,
		Action:     action.ActionWrite,
		Specifiers: specifier.SpecifierGroup{},
	}.Create()
	expectedG1R1WritePolicy := g1R1WritePolicy
	expectedG1R1WritePolicy.Specifiers = specifier.SpecifierGroup{Specifiers: []specifier.Specifier{envRoot}}

	testCases := []struct {
		name       string
		subject    subject.Subject
		action     action.Action
		resource   resource.Resource
		specifiers []specifier.Specifier
		expected   []policy.Policy
	}{
		{"all policies", subject.Subject{}, "", resource.Resource{}, []specifier.Specifier{}, []policy.Policy{expectedG1R1ReadPolicy, expectedG1R1WritePolicy, expectedG2R1ProdReadPolicy}},
		{"policy by action READ", subject.Subject{}, action.ActionRead, resource.Resource{}, []specifier.Specifier{}, []policy.Policy{expectedG1R1ReadPolicy, expectedG2R1ProdReadPolicy}},
		{"policy by action WRITE", subject.Subject{}, action.ActionWrite, resource.Resource{}, []specifier.Specifier{}, []policy.Policy{expectedG1R1WritePolicy}},
		{"policy by resource r1", subject.Subject{}, "", r1, []specifier.Specifier{}, []policy.Policy{expectedG1R1ReadPolicy, expectedG1R1WritePolicy, expectedG2R1ProdReadPolicy}},
		{"policy by subject g1", g1, "", resource.Resource{}, []specifier.Specifier{}, []policy.Policy{expectedG1R1ReadPolicy, expectedG1R1WritePolicy}},
		{"policy by specifier Env:*", subject.Subject{}, "", resource.Resource{}, []specifier.Specifier{envRoot}, []policy.Policy{expectedG1R1ReadPolicy, expectedG1R1WritePolicy}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := policy.Policy{
				Subject:    tc.subject,
				Resource:   tc.resource,
				Action:     tc.action,
				Specifiers: specifier.SpecifierGroup{Specifiers: tc.specifiers},
			}.Get()

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
				return
			}

			slices.SortFunc(result, func(a, b policy.Policy) int {
				return strings.Compare(a.Id, b.Id)
			})
			slices.SortFunc(tc.expected, func(a, b policy.Policy) int {
				return strings.Compare(a.Id, b.Id)
			})

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("For %s, expected %v, but got %v", tc.name, tc.expected, result)
			}
		})
	}
}
