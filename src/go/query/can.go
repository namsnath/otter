package query

import (
	"fmt"
	"log/slog"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

// CanQueryBuilder holds the state of the query as it is being built.
// Can <subject> Perform <action> On <resource> With <specifiers>
type CanQueryBuilder struct {
	subject    subject.Subject
	action     action.Action
	resource   resource.Resource
	specifiers map[string]string
}

// `Can` initializes a new QueryBuilder and sets the Subject.
func Can(s subject.Subject) *CanQueryBuilder {
	return &CanQueryBuilder{
		subject: s,
	}
}

// Perform sets the Action on the QueryBuilder.
func (qb *CanQueryBuilder) Perform(a action.Action) *CanQueryBuilder {
	qb.action = a
	return qb // Return the receiver struct
}

// On sets the Resource on the QueryBuilder.
func (qb *CanQueryBuilder) On(r resource.Resource) *CanQueryBuilder {
	qb.resource = r
	return qb // Return the receiver struct
}

// With sets the SpecifierGroup on the QueryBuilder.
func (qb *CanQueryBuilder) With(sg specifier.SpecifierGroup) *CanQueryBuilder {
	qb.specifiers = sg.AsMap()
	return qb // Return the receiver struct
}

func (qb *CanQueryBuilder) Validate() (*CanQueryBuilder, error) {
	if qb.subject == (subject.Subject{}) || qb.action == "" || qb.resource == (resource.Resource{}) {
		return nil, fmt.Errorf("incomplete Can query: subject, action, and resource must be set")
	}
	return qb, nil
}

func (qb *CanQueryBuilder) Query() (bool, error) {
	qb, ok := qb.Validate()
	if ok != nil {
		return false, ok
	}

	query := `
		MATCH (specifier:Specifier)
		WHERE specifier.key <> "*"
		WITH DISTINCT specifier.key as AllKeys
		UNWIND AllKeys AS k
		WITH apoc.map.mergeList(collect(apoc.map.setKey({}, k, "*"))) AS AllSpecMap
		WITH apoc.map.merge(AllSpecMap, $specifiers) AS NormalizedSpecifiers

		UNWIND keys(NormalizedSpecifiers) AS k
		WITH NormalizedSpecifiers, k, NormalizedSpecifiers[k] AS v

		MATCH (s:Specifier)
		WHERE s.key = k AND s.value = v

		MATCH (p:Policy)-[:$($action)]->(ps:Specifier)<-[:CHILD_OF*0..]-(s)

		MATCH (subject:Subject {name: $subject})-[:CHILD_OF*0..]->(parents:Subject)-[:HAS_POLICY]->(p)
		MATCH (resource:Resource {name: $resource})-[:CHILD_OF*0..]->(:Resource)-[:HAS_POLICY]->(p)


		// Aggregate by Policy and count how many *distinct* keys were matched
		WITH p, count(DISTINCT s.key) AS matches, size(keys(NormalizedSpecifiers)) AS requiredMatches, NormalizedSpecifiers

		// The Policy is valid only if it matched EVERY key in the input
		WHERE matches = requiredMatches

		RETURN p IS NOT NULL AS CanDo
	`

	params := map[string]any{
		"subject":    qb.subject.Name,
		"resource":   qb.resource.Name,
		"action":     string(qb.action),
		"specifiers": qb.specifiers,
	}

	var canDoRaw interface{}

	queryResult := db.ExecuteQuery(query, params)
	if len(queryResult.Records) == 0 {
		return false, nil
	}

	canDoRaw, hasVal := queryResult.Records[0].Get("CanDo")
	if !hasVal {
		return false, nil
	}

	canDo := canDoRaw.(bool)

	slog.Info("Can",
		"subject", qb.subject,
		"action", qb.action,
		"resource", qb.resource,
		"specifiers", qb.specifiers,
		"canDo", canDo,
		"duration", queryResult.Summary.ResultAvailableAfter(),
	)

	return canDo, nil
}
