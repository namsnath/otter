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
type WhoCanQueryBuilder struct {
	action     action.Action
	resource   resource.Resource
	specifiers map[string]string
}

// `WhoCan` initializes a new QueryBuilder and sets the Subject.
func WhoCan(a action.Action) *WhoCanQueryBuilder {
	return &WhoCanQueryBuilder{
		action: a,
	}
}

// On sets the Resource on the QueryBuilder.
func (qb *WhoCanQueryBuilder) On(r resource.Resource) *WhoCanQueryBuilder {
	qb.resource = r
	return qb // Return the receiver struct
}

// With sets the SpecifierGroup on the QueryBuilder.
func (qb *WhoCanQueryBuilder) With(sg specifier.SpecifierGroup) *WhoCanQueryBuilder {
	qb.specifiers = sg.AsMap()
	return qb // Return the receiver struct
}

func (qb *WhoCanQueryBuilder) Validate() (*WhoCanQueryBuilder, error) {
	if qb.action == "" || qb.resource == (resource.Resource{}) {
		return nil, fmt.Errorf("incomplete WhoCan query: action and resource must be set")
	}
	return qb, nil
}

func (qb *WhoCanQueryBuilder) Query() ([]subject.Subject, error) {
	qb, ok := qb.Validate()
	if ok != nil {
		return []subject.Subject{}, ok
	}

	// TODO: Improve the specifiers conditional to check for supersets of the defined specifiers
	query := `
		MATCH (s:Subject)-[:CHILD_OF*0..]->(:Subject)-[rel:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r:Resource {name: $resourceName})
		WHERE properties(rel) = $specifiers
		RETURN DISTINCT s.name AS subject, s.type AS subjectType
	`
	params := map[string]any{
		"resourceName": qb.resource.Name,
		"action":       string(qb.action),
		"specifiers":   qb.specifiers,
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
		"duration", result.Summary.ResultAvailableAfter())

	return subjects, nil
}

// func AllSubjectsThatCanDo(resource *resource.Resource, action action.Action, specifiers *specifier.SpecifierGroup) []subject.Subject {
// 	edgeProps := map[string]string{}
// 	if specifiers != nil {
// 		edgeProps = specifiers.AsMap()
// 	}

// 	result := db.ExecuteQuery(`
// 		MATCH (s:Subject)-[:CHILD_OF*0..]->(:Subject)-[rel:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r:Resource {name: $resourceName})
// 		WHERE properties(rel) = $edgeProps
// 		RETURN DISTINCT s.name AS subject, s.type AS subjectType
// 		`,
// 		map[string]any{
// 			"resourceName": resource.Name,
// 			"action":       string(action),
// 			"edgeProps":    edgeProps,
// 		},
// 	)

// 	subjects := make([]subject.Subject, 0, len(result.Records))
// 	for _, record := range result.Records {
// 		nameVal, nameOk := record.Get("subject")
// 		typeVal, typeOk := record.Get("subjectType")
// 		if nameOk && typeOk {
// 			if nameStr, nameIsStr := nameVal.(string); nameIsStr {
// 				subject := subject.Subject{Name: nameStr, Type: subject.SubjectTypeFromString(typeVal.(string))}
// 				subjects = append(subjects, subject)
// 			}
// 		}
// 	}

// 	slog.Info("AllSubjectsThatCanDo",
// 		"resource", resource,
// 		"subjects", subjects,
// 		"specifiers", edgeProps,
// 		"duration", result.Summary.ResultAvailableAfter())

// 	return subjects
// }
