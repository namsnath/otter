package policy

import "github.com/namsnath/otter/db"

func (policy Policy) Create() Policy {
	query := `
		MATCH (subject:Subject {name: $subjectName})
		MATCH (resource:Resource {name: $resourceName})
		CREATE (policy:Policy {id: randomUUID()})
		CREATE (subject)-[:HAS_POLICY]->(policy)<-[:HAS_POLICY]-(resource)

		WITH policy
		UNWIND keys($specifiers) AS k
		MATCH (specifier:Specifier {key: k, value: $specifiers[k]})
		CREATE (policy)-[e:$($action)]->(specifier)
	`
	params := map[string]any{
		"subjectName":  policy.Subject.Name,
		"resourceName": policy.Resource.Name,
		"action":       string(policy.Action),
		"specifiers":   policy.Specifiers.AsMap(),
	}

	if len(policy.Specifiers.Specifiers) == 0 {
		params["specifiers"] = map[string]string{"*": "*"}
	}

	db.ExecuteQuery(query, params)

	return policy
}
