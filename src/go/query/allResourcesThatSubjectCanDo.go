package query

import (
	"log/slog"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

func AllResourcesThatSubjectCanDo(subject *subject.Subject, action action.Action, specifiers *specifier.SpecifierGroup) []resource.Resource {
	edgeProps := map[string]string{}
	if specifiers != nil {
		edgeProps = specifiers.AsMap()
	}

	result := db.ExecuteQuery(`
		MATCH (s:Subject {name: $subjectName})-[:CHILD_OF*0..]->(:Subject)-[rel:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r:Resource)
		WHERE properties(rel) = $edgeProps
		RETURN DISTINCT r.name AS resource
		`,
		map[string]any{
			"subjectName": subject.Name,
			"action":      string(action),
			"edgeProps":   edgeProps,
		},
	)

	resources := make([]resource.Resource, 0, len(result.Records))
	for _, record := range result.Records {
		nameVal, nameOk := record.Get("resource")
		if nameOk {
			if nameStr, nameIsStr := nameVal.(string); nameIsStr {
				resource := resource.Resource{Name: nameStr}
				resources = append(resources, resource)
			}
		}
	}

	slog.Info("AllResourcesThatSubjectCanDo",
		"subject", subject,
		"resources", resources,
		"specifiers", edgeProps,
		"duration", result.Summary.ResultAvailableAfter())

	return resources
}
