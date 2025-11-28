package policy

import "github.com/namsnath/otter/db"

func (policy Policy) Create() Policy {
	// TODO: Change this so that the same policy is used for all specifiers in the group. Unwind and create only policy -> specifier relationships in loop
	query := `
		MATCH (subject:Subject {name: $subjectName})
		MATCH (resource:Resource {name: $resourceName})
		MATCH (specifier:Specifier {key: $specifierKey, value: $specifierValue})
		CREATE (policy:Policy {id: randomUUID()})
		CREATE (subject)-[:HAS_POLICY]->(policy)<-[:HAS_POLICY]-(resource)
		CREATE (policy)-[e:$($action)]->(specifier)
	`
	params := map[string]any{
		"subjectName":  policy.Subject.Name,
		"resourceName": policy.Resource.Name,
		"action":       string(policy.Action),
	}

	if len(policy.Specifiers.Specifiers) == 0 {
		// No specifiers, create policy using the root specifier
		params["specifierKey"] = "*"
		params["specifierValue"] = "*"
		db.ExecuteQuery(query, params)
	} else {
		for _, specifier := range policy.Specifiers.Specifiers {
			params["specifierKey"] = specifier.Key
			params["specifierValue"] = specifier.Value
			db.ExecuteQuery(query, params)
		}
	}

	return policy
}
