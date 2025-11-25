package query

import (
	"log/slog"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

func AllSubjectsThatCanDo(resource *resource.Resource, action action.Action, specifier *specifier.Specifier) []subject.Subject {
	// Subject s is a child of any Subject that can READ a Resource that is a parent of any Resouce r,
	// where r is the input Resource
	result := db.ExecuteQuery(`
		MATCH (s:Subject)-[:CHILD_OF*0..]->(:Subject)-[:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r:Resource {name: $resourceName})
		RETURN DISTINCT s.name AS subject, s.type AS subjectType
		`,
		map[string]any{
			"resourceName": resource.Name,
			"action":       string(action),
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
		"duration", result.Summary.ResultAvailableAfter())

	return subjects
}
