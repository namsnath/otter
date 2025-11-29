package policy

import "github.com/namsnath/otter/db"

func (policy Policy) Create() Policy {
	query := `
		MATCH (specifier:Specifier)
		WHERE specifier.key <> "*"
		WITH DISTINCT specifier.key as AllKeys
		UNWIND AllKeys AS k
		WITH apoc.map.mergeList(collect(apoc.map.setKey({}, k, "*"))) AS AllSpecMap
		WITH apoc.map.merge(AllSpecMap, $specifiers) AS NormalizedSpecifiers

		MATCH (subject:Subject {name: $subjectName})
		MATCH (resource:Resource {name: $resourceName})
		CREATE (policy:Policy {id: randomUUID()})
		CREATE (subject)-[:HAS_POLICY]->(policy)<-[:HAS_POLICY]-(resource)

		WITH policy, NormalizedSpecifiers
		UNWIND keys(NormalizedSpecifiers) AS k
		MATCH (specifier:Specifier {key: k, value: NormalizedSpecifiers[k]})
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
