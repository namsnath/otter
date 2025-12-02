package policy

import (
	"github.com/namsnath/otter/db"
)

func (policy Policy) Create() (Policy, error) {
	query := `
		MATCH (specifier:Specifier)
		WHERE specifier.key <> "*"
		WITH collect(DISTINCT specifier.key) AS allKeys
		WITH reduce(specMap = $specifiers, k IN allKeys |
			CASE WHEN NOT k IN keys(specMap) THEN apoc.map.setKey(specMap, k, "*") ELSE specMap END
		) AS normalizedSpecifiers

		MATCH (subject:Subject {name: $subjectName})
		MATCH (resource:Resource {name: $resourceName})
		CREATE (policy:Policy {id: randomUUID()})
		CREATE (subject)-[:HAS_POLICY]->(policy)<-[:HAS_POLICY]-(resource)

		WITH policy, normalizedSpecifiers
		UNWIND keys(normalizedSpecifiers) AS k
		MATCH (specifier:Specifier {key: k, value: normalizedSpecifiers[k]})
		CREATE (policy)-[e:$($action)]->(specifier)

		RETURN DISTINCT policy.id as PolicyId
	`

	params := map[string]any{
		"subjectName":  policy.Subject.Name,
		"resourceName": policy.Resource.Name,
		"action":       string(policy.Action),
		"specifiers":   policy.Specifiers.AsMap(),
	}

	result := db.ExecuteQuery(query, params)
	if len(result.Records) == 0 {
		return Policy{}, nil
	}

	policyId := result.Records[0].AsMap()["PolicyId"].(string)

	policy.Id = policyId
	return policy, nil
}
