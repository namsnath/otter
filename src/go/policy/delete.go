package policy

import (
	"github.com/namsnath/otter/db"
)

func (policy Policy) Delete() error {
	if policy.Id == "" {
		return ErrPolicyIDRequired
	}

	query := `
		MATCH (p:Policy {id: $policyId})
		DETACH DELETE p
	`

	params := map[string]any{
		"policyId": policy.Id,
	}

	db.ExecuteQuery(query, params)
	return nil
}
