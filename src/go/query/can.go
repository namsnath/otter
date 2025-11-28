package query

import (
	"fmt"
	"log/slog"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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

	// TODO: Improve the specifiers conditional to check for supersets of the defined specifiers
	// TODO: Can we unwind the specifiers in the query itself instead of looping here?
	query := `
		MATCH (subject:Subject {name: $subject})
		MATCH (resource:Resource {name: $resource})
		MATCH (subject)-[:CHILD_OF*0..]->(:Subject)-[:HAS_POLICY]->(policy:Policy)<-[:HAS_POLICY]-(:Resource)<-[:CHILD_OF*0..]-(resource)
		RETURN EXISTS {
			(policy)-[rel:$($action)]->(:Specifier)<-[:CHILD_OF*0..]-(specifier:Specifier {key: $specifierKey, value: $specifierValue})
		} as CanDo
	`
	params := map[string]any{
		"subject":  qb.subject.Name,
		"resource": qb.resource.Name,
		"action":   string(qb.action),
	}

	var canDoRaw interface{}
	var canDo, hasVal bool
	var queryResult *neo4j.EagerResult

	if len(qb.specifiers) == 0 {
		params["specifierKey"] = "*"
		params["specifierValue"] = "*"
		queryResult = db.ExecuteQuery(query, params)
		if len(queryResult.Records) == 0 {
			return false, nil
		}

		canDoRaw, hasVal = queryResult.Records[0].Get("CanDo")
		if !hasVal {
			return false, nil
		}

		canDo = canDoRaw.(bool)
	} else {
		canDo = true
		for key, value := range qb.specifiers {
			params["specifierKey"] = key
			params["specifierValue"] = value
			queryResult = db.ExecuteQuery(query, params)
			if len(queryResult.Records) == 0 {
				return false, nil
			}

			canDoRaw, hasVal = queryResult.Records[0].Get("CanDo")
			if !hasVal || !canDoRaw.(bool) {
				canDo = false
				break
			}
		}
	}

	slog.Info("Can",
		"subject", qb.subject,
		"action", qb.action,
		"resource", qb.resource,
		"specifiers", qb.specifiers,
		"canDo", canDo,
		// "duration", result.Summary.ResultAvailableAfter(),
	)

	return canDo, nil
}
