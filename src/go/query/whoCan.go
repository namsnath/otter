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

// WhoCanQueryBuilder holds the state of the query as it is being built.
// WhoCan <action> On <resource> With <specifiers>
type WhoCanQueryBuilder struct {
	action     action.Action
	resource   resource.Resource
	specifiers map[string]string
}

// `WhoCan` initializes a new QueryBuilder and sets the Subject.
func WhoCan(a action.Action) WhoCanQueryBuilder {
	return WhoCanQueryBuilder{
		action: a,
	}
}

// On sets the Resource on the QueryBuilder.
func (qb WhoCanQueryBuilder) On(r resource.Resource) WhoCanQueryBuilder {
	qb.resource = r
	return qb
}

// With sets the SpecifierGroup on the QueryBuilder.
func (qb WhoCanQueryBuilder) With(sg specifier.SpecifierGroup) WhoCanQueryBuilder {
	qb.specifiers = sg.AsMap()
	return qb
}

func (qb WhoCanQueryBuilder) Validate() (WhoCanQueryBuilder, error) {
	if qb.action == "" || qb.resource == (resource.Resource{}) {
		return WhoCanQueryBuilder{}, fmt.Errorf("incomplete WhoCan query: action and resource must be set")
	}
	return qb, nil
}

func (qb WhoCanQueryBuilder) Query() ([]subject.Subject, error) {
	qb, ok := qb.Validate()
	if ok != nil {
		return []subject.Subject{}, ok
	}

	query := `
		MATCH (specifier:Specifier)
		WHERE specifier.key <> "*"
		WITH collect(DISTINCT specifier.key) AS allKeys
		WITH reduce(specMap = $specifiers, k IN allKeys |
			CASE WHEN NOT k IN keys(specMap) THEN apoc.map.setKey(specMap, k, "*") ELSE specMap END
		) AS normalizedSpecifiers

		UNWIND keys(normalizedSpecifiers) AS k
		WITH normalizedSpecifiers, k, normalizedSpecifiers[k] AS v

		MATCH (s:Specifier)
		WHERE s.key = k AND s.value = v

		MATCH (p:Policy)-[:$($action)]->(ps:Specifier)<-[:CHILD_OF*0..]-(s)
		MATCH (resource:Resource {name: $resource})-[:CHILD_OF*0..]->(:Resource)-[:HAS_POLICY]->(p)

		WITH p, count(DISTINCT s.key) AS matches, size(keys(normalizedSpecifiers)) AS requiredMatches, normalizedSpecifiers
		WHERE matches = requiredMatches

		MATCH (subject:Subject)-[:CHILD_OF*0..]->(:Subject)-[:HAS_POLICY]->(p)

		RETURN DISTINCT subject.name AS subject, subject.type AS subjectType
	`

	params := map[string]any{
		"resource":   qb.resource.Name,
		"action":     string(qb.action),
		"specifiers": qb.specifiers,
	}

	result := db.ExecuteQuery(query, params)

	subjects := make([]subject.Subject, 0, len(result.Records))
	for _, record := range result.Records {
		nameVal, nameOk := record.Get("subject")
		typeVal, typeOk := record.Get("subjectType")
		if nameOk && typeOk {
			if nameStr, nameIsStr := nameVal.(string); nameIsStr {
				subject := subject.Subject{Name: nameStr, Type: subject.SubjectTypeFromString(typeVal.(string))}
				subjects = append(subjects, subject)
			}
		}
	}

	slog.Info("WhoCan",
		"action", qb.action,
		"resource", qb.resource,
		"specifiers", qb.specifiers,
		"subjects", subjects,
		"duration", result.Summary.ResultAvailableAfter(),
		"rows", len(result.Records),
	)

	return subjects, nil
}
