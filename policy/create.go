package policy

import "github.com/namsnath/gatekeeper/db"

func (policy Policy) Create() {
	// TODO: Handle specifiers, need to serialize the map as properties on the relationship
	db.ExecuteQuery(`
		MATCH (s:Subject {name: $subjectName})
		MATCH (r:Resource {name: $resourceName})
		CREATE (s)-[:$($action)]->(r)
		`,
		map[string]any{
			"subjectName":  policy.Subject.Name,
			"resourceName": policy.Resource.Name,
			"action":       string(policy.Action),
		},
	)
}
