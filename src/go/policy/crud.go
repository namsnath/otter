package policy

import (
	"github.com/namsnath/otter/db"
)

func (policy Policy) Create() {
	db.ExecuteQuery(`
		MATCH (s:Subject {name: $subjectName})
		MATCH (r:Resource {name: $resourceName})
		CREATE (s)-[e:$($action)]->(r)
		SET e = $edgeProps
		`,
		map[string]any{
			"subjectName":  policy.Subject.Name,
			"resourceName": policy.Resource.Name,
			"action":       string(policy.Action),
			"edgeProps":    policy.Specifiers.AsMap(),
		},
	)
}
