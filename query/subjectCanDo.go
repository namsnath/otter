package query

import (
	"log/slog"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

func SubjectCanDo(subject *subject.Subject, action action.Action, resource *resource.Resource, specifier *specifier.Specifier) bool {
	result := db.ExecuteQuery(`
		MATCH (s:Subject {name: $subject})
		MATCH (r:Resource {name: $resource})
		RETURN EXISTS { (s)-[:CHILD_OF*0..]->(:Subject)-[rel:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r) } as CanDo
		`,
		map[string]any{
			"subject":  subject.Name,
			"resource": resource.Name,
			"action":   string(action),
		},
	)

	if len(result.Records) == 0 {
		return false
	}

	val, ok := result.Records[0].Get("CanDo")
	if !ok {
		return false
	}

	slog.Info("SubjectCanDo",
		"resource", resource,
		"canDo", val,
		"duration", result.Summary.ResultAvailableAfter())

	return val.(bool)
}
