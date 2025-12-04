package query_test

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/policy"
	"github.com/namsnath/otter/query"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

func deletePolicies() {
	db.ExecuteQuery("MATCH (p:Policy) DETACH DELETE p", nil)
}

func TestHowCanQuery(t *testing.T) {
	ctx, container := db.TestContainer()
	// Ensure the container is terminated after the test finishes
	defer func() {
		container.Terminate(ctx)
	}()

	query.DeleteEverything()
	query.SetupIndexes()

	g1 := subject.Subject{Name: "Group1", Type: subject.SubjectTypeGroup}.Create()
	p1 := subject.Subject{Name: "Principal1", Type: subject.SubjectTypePrincipal}.CreateAsChildOf(g1)
	rRoot := resource.Resource{Name: "_"}.Create()
	r1 := resource.Resource{Name: "Resource1"}.CreateAsChildOf(rRoot)
	rootSpecifier := specifier.NewSpecifier("*", "*").Create()
	envRoot, _ := specifier.NewSpecifier("Env", "*").CreateAsChildOf(rootSpecifier)
	envProd, _ := specifier.NewSpecifier("Env", "prod").CreateAsChildOf(envRoot)
	envDev, _ := specifier.NewSpecifier("Env", "dev").CreateAsChildOf(envRoot)
	roleRoot, _ := specifier.NewSpecifier("Role", "*").CreateAsChildOf(rootSpecifier)
	roleAdmin, _ := specifier.NewSpecifier("Role", "admin").CreateAsChildOf(roleRoot)
	roleUser, _ := specifier.NewSpecifier("Role", "user").CreateAsChildOf(roleAdmin)

	p1R1AdminDevPolicy := policy.Policy{
		Subject:    p1,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleAdmin, envDev}},
	}
	p1R1UserProdPolicy := policy.Policy{
		Subject:    p1,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleUser, envProd}},
	}
	g1R1UserDevPolicy := policy.Policy{
		Subject:    g1,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleUser, envDev}},
	}
	p1RRootAdminDevPolicy := policy.Policy{
		Subject:    p1,
		Resource:   rRoot,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleAdmin, envDev}},
	}
	p1RRootUserProdPolicy := policy.Policy{
		Subject:    p1,
		Resource:   rRoot,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleUser, envProd}},
	}
	p1RRootAllEnvAdminPolicy := policy.Policy{
		Subject:    p1,
		Resource:   rRoot,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleAdmin}},
	}
	g1R1AllEnvAdminPolicy := policy.Policy{
		Subject:    g1,
		Resource:   r1,
		Action:     action.ActionRead,
		Specifiers: specifier.SpecifierGroup{Specifiers: []specifier.Specifier{roleAdmin}},
	}

	testCases := []struct {
		name       string
		policies   []policy.Policy
		subject    subject.Subject
		action     action.Action
		resource   resource.Resource
		specifiers []specifier.Specifier
		expected   []specifier.SpecifierGroup
	}{
		{
			"specific policies, subject inheritance, no specifier filter",
			[]policy.Policy{p1R1AdminDevPolicy, p1R1UserProdPolicy, g1R1UserDevPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{roleAdmin, envDev}},
				{Specifiers: []specifier.Specifier{roleUser, envProd}},
				{Specifiers: []specifier.Specifier{roleUser, envDev}},
			},
		},
		{
			"specific policies, subject inheritance, with specifier filter",
			[]policy.Policy{p1R1AdminDevPolicy, p1R1UserProdPolicy, g1R1UserDevPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{roleUser},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{envProd}},
				{Specifiers: []specifier.Specifier{envDev}},
			},
		},
		{
			"specific policies, resource inheritance, no specifier filter",
			[]policy.Policy{p1RRootAdminDevPolicy, p1RRootUserProdPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{roleAdmin, envDev}},
				{Specifiers: []specifier.Specifier{roleUser, envDev}},
				{Specifiers: []specifier.Specifier{roleUser, envProd}},
			},
		},
		{
			"specific policies, resource inheritance, with specifier filter",
			[]policy.Policy{p1RRootAdminDevPolicy, p1RRootUserProdPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{roleUser},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{envProd}},
				{Specifiers: []specifier.Specifier{envDev}},
			},
		},
		{
			"wildcard policies, resource inheritance, no specifier filter",
			[]policy.Policy{p1RRootAllEnvAdminPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{roleAdmin, envRoot}},
				{Specifiers: []specifier.Specifier{roleUser, envRoot}},
			},
		},
		{
			"wildcard policies, resource inheritance, with specific specifier filter",
			[]policy.Policy{p1RRootAllEnvAdminPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{roleUser},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{envRoot}},
			},
		},
		{
			"wildcard policies, resource inheritance, with wildcard specifier filter",
			[]policy.Policy{p1RRootAllEnvAdminPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{envRoot},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{roleUser}},
				{Specifiers: []specifier.Specifier{roleAdmin}},
			},
		},

		{
			"wildcard policies, subject inheritance, no specifier filter",
			[]policy.Policy{g1R1AllEnvAdminPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{roleAdmin, envRoot}},
				{Specifiers: []specifier.Specifier{roleUser, envRoot}},
			},
		},
		{
			"wildcard policies, subject inheritance, with specific specifier filter",
			[]policy.Policy{g1R1AllEnvAdminPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{roleUser},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{envRoot}},
			},
		},
		{
			"wildcard policies, subject inheritance, with wildcard specifier filter",
			[]policy.Policy{g1R1AllEnvAdminPolicy},
			p1, action.ActionRead, r1, []specifier.Specifier{envRoot},
			[]specifier.SpecifierGroup{
				{Specifiers: []specifier.Specifier{roleUser}},
				{Specifiers: []specifier.Specifier{roleAdmin}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deletePolicies()
			for _, p := range tc.policies {
				_, err := p.Create()
				if err != nil {
					slog.Error("Error creating policy", "policy", p, "error", err)
				}
			}

			result, err := query.HowCan(tc.subject).Perform(tc.action).On(tc.resource).With(specifier.SpecifierGroup{Specifiers: tc.specifiers}).Query()

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
				return
			}

			resultList := []string{}
			expectedList := []string{}

			for _, sg := range result {
				groupList := []string{}
				for _, specifier := range sg.Specifiers {
					groupList = append(groupList, fmt.Sprintf("%s=%s", specifier.Key, specifier.Value))
				}
				slices.Sort(groupList)
				resultList = append(resultList, strings.Join(groupList, ","))
			}

			for _, sg := range tc.expected {
				groupList := []string{}
				for _, specifier := range sg.Specifiers {
					groupList = append(groupList, fmt.Sprintf("%s=%s", specifier.Key, specifier.Value))
				}
				slices.Sort(groupList)
				expectedList = append(expectedList, strings.Join(groupList, ","))
			}

			slices.Sort(resultList)
			slices.Sort(expectedList)

			// slices.SortFunc(result, func(a, b specifier.SpecifierGroup) int {
			// 	return strings.Compare(a.Id, b.Id)
			// })
			// slices.SortFunc(tc.expected, func(a, b specifier.SpecifierGroup) int {
			// 	return strings.Compare(a.Id, b.Id)
			// })

			if !reflect.DeepEqual(resultList, expectedList) {
				t.Errorf("For %s, expected %v, but got %v", tc.name, expectedList, resultList)
			}
		})
	}
}
