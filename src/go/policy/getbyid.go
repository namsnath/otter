package policy

import (
	"fmt"

	"github.com/namsnath/otter/db"
)

func (policy Policy) GetById() (Policy, error) {
	if policy.Id == "" {
		return Policy{}, fmt.Errorf("policy Id should be specified")
	}

	query := `
		MATCH (policy:Policy {id: $policyId})
		MATCH (subject:Subject)-[:HAS_POLICY]->(policy)
		MATCH (resource:Resource)-[:HAS_POLICY]->(policy)
		MATCH (specifier:Specifier)<-[rel]-(policy)

		RETURN DISTINCT policy.id as policyId, subject, resource, type(rel) AS action, collect(specifier) AS specifiers
	`

	params := map[string]any{
		"policyId": policy.Id,
	}

	result := db.ExecuteQuery(query, params)

	if len(result.Records) == 0 {
		return Policy{}, nil
	}

	resultPolicy, err := ProcessPolicyRecord(result.Records[0])
	if err != nil {
		return Policy{}, err
	}
	return resultPolicy, nil
}
