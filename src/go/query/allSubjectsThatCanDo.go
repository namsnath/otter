package query

import (
	"log/slog"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

func AllSubjectsThatCanDo(resource *resource.Resource, action action.Action, specifiers *specifier.SpecifierGroup) []subject.Subject {
	edgeProps := map[string]string{}
	if specifiers != nil {
		edgeProps = specifiers.AsMap()
	}

	result := db.ExecuteQuery(`
		MATCH (s:Subject)-[:CHILD_OF*0..]->(:Subject)-[rel:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r:Resource {name: $resourceName})
		WHERE properties(rel) = $edgeProps
		RETURN DISTINCT s.name AS subject, s.type AS subjectType
		`,
		map[string]any{
			"resourceName": resource.Name,
			"action":       string(action),
			"edgeProps":    edgeProps,
		},
	)

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

	slog.Info("AllSubjectsThatCanDo",
		"resource", resource,
		"subjects", subjects,
		"specifiers", edgeProps,
		"duration", result.Summary.ResultAvailableAfter())

	return subjects
}
