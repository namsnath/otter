package policy

import (
	"log/slog"

	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/subject"
)

func (policy Policy) Get() ([]Policy, error) {
	query := `
	CALL () {
		// --- BRANCH A: Specifiers is NULL ---
		// Fetch ALL policies and their specifiers
		WITH $specifiers AS inputMap
		WHERE inputMap IS NULL

		MATCH (p:Policy)-[r]->(s:Specifier)
		WHERE CASE
			WHEN $action IS NOT NULL
			THEN type(r) = $action
			ELSE TRUE
		END
		RETURN p, type(r) as action, collect(s) as specifiers

		UNION

		// --- BRANCH B: Specifiers is not NULL ---
		WITH $specifiers AS inputMap
		WHERE inputMap IS NOT NULL

		// B1. Fetch all DB keys to handle implicit wildcards
		CALL () {
			MATCH (s:Specifier) WHERE s.key <> "*"
			RETURN collect(DISTINCT s.key) AS allSpecifierKeys
		}

		// B2. Normalize the map (Fill missing keys with '*')
		WITH inputMap,
			reduce(acc = inputMap, k IN allSpecifierKeys |
				CASE
					WHEN k IN keys(inputMap) THEN acc
					ELSE apoc.map.setKey(acc, k, "*")
				END
			) AS normalizedSpecifiers

		UNWIND keys(normalizedSpecifiers) AS k
		WITH normalizedSpecifiers, k, normalizedSpecifiers[k] AS v

		MATCH (s:Specifier)
		WHERE s.key = k AND s.value = v
		// WHERE s.key = k AND (s.value = v OR s.value = '*')

		MATCH (p:Policy)-[r]->(s)
		WHERE CASE
			WHEN $action IS NOT NULL
			THEN type(r) = $action
			ELSE TRUE
		END

		// Ensure that ALL keys in the normalized map are matched
		WITH p, r, collect(s) AS specifiers,
			count(DISTINCT s.key) AS matches,
			size(keys(normalizedSpecifiers)) AS requiredMatches

		WHERE matches = requiredMatches
		RETURN p, type(r) as action, specifiers
	}

	MATCH (subject:Subject)-[:HAS_POLICY]->(p)
		WHERE $subject IS NULL OR subject.name = $subject

	MATCH (resource:Resource)-[:HAS_POLICY]->(p)
		WHERE $resource IS NULL OR resource.name = $resource

	RETURN
		p.id AS policyId,
		action,
		specifiers,
		subject,
		resource
	`

	params := map[string]any{
		"subject":    nil,
		"resource":   nil,
		"action":     nil,
		"specifiers": nil,
	}

	if policy.Action != "" {
		params["action"] = string(policy.Action)
	}

	if policy.Subject != (subject.Subject{}) {
		params["subject"] = policy.Subject.Name
	}

	if policy.Resource != (resource.Resource{}) {
		params["resource"] = policy.Resource.Name
	}

	if len(policy.Specifiers.Specifiers) > 0 {
		params["specifiers"] = policy.Specifiers.AsMap()
	}

	result := db.ExecuteQuery(query, params)

	slog.Info(
		"Policy.Get",
		"subject", policy.Subject,
		"resource", policy.Resource,
		"action", policy.Action,
		"specifiers", policy.Specifiers,
		"rows", len(result.Records),
		"duration", result.Summary.ResultAvailableAfter(),
	)

	if len(result.Records) == 0 {
		return []Policy{}, nil
	}

	policies := []Policy{}

	for _, record := range result.Records {
		policy, err := ProcessPolicyRecord(record)
		if err != nil {
			return []Policy{}, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}
