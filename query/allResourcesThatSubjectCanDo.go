package query

import (
	"log/slog"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

func AllResourcesThatSubjectCanDo(subject *subject.Subject, action action.Action, specifier *specifier.Specifier) []resource.Resource {
	// Subject s is a child of any Subject that can ACTION a Resource that is a parent of any Resouce r
	result := db.ExecuteQuery(`
		MATCH (s:Subject {name: $subjectName})-[:CHILD_OF*0..]->(:Subject)-[:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r:Resource)
		RETURN DISTINCT r.name AS resource
		`,
		map[string]any{
			"subjectName": subject.Name,
			"action":      string(action),
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
		"duration", result.Summary.ResultAvailableAfter())

	return resources
}
