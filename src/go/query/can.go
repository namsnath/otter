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

	// TODO: Improve the specifiers conditional to check for supersets of the defined specifiers
	query := `
		MATCH (s:Subject {name: $subject})
		MATCH (r:Resource {name: $resource})
		RETURN EXISTS {
			(s)-[:CHILD_OF*0..]->(:Subject)-[rel:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r)
			WHERE properties(rel) = $edgeProps
		} as CanDo
	`
	params := map[string]any{
		"subject":   qb.subject.Name,
		"resource":  qb.resource.Name,
		"action":    string(qb.action),
		"edgeProps": qb.specifiers,
	}
	result := db.ExecuteQuery(query, params)

	if len(result.Records) == 0 {
		return false, nil
	}

	canDo, hasVal := result.Records[0].Get("CanDo")
	if !hasVal {
		return false, nil
	}

	slog.Info("Can",
		"subject", qb.subject,
		"action", qb.action,
		"resource", qb.resource,
		"specifiers", qb.specifiers,
		"canDo", canDo,
		"duration", result.Summary.ResultAvailableAfter())

	return canDo.(bool), nil
}
