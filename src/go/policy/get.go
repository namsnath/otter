package policy

import (
	"log/slog"

	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/subject"
)

func (policy Policy) Get() ([]Policy, error) {
	// query := `
	// 	WITH coalesce($specifiers, {}) AS inputMap

	// 	CALL () {
	// 		MATCH (s:Specifier) WHERE s.key <> "*"
	// 		RETURN collect(DISTINCT s.key) AS allDbKeys
	// 	}

	// 	WITH inputMap,
	// 		reduce(acc = inputMap, k IN allDbKeys |
	// 			CASE
	// 				WHEN k IN keys(inputMap) THEN acc
	// 				ELSE apoc.map.setKey(acc, k, "*")
	// 			END
	// 			) AS normalizedSpecifiers

	// 	UNWIND keys(normalizedSpecifiers) AS k
	// 	WITH normalizedSpecifiers, k, normalizedSpecifiers[k] AS v

	// 	MATCH (s:Specifier)
	// 		WHERE s.key = k AND s.value = v

	// 	MATCH (specifier:Specifier)<-[rel:$($action)]-(policy:Policy)

	// 	MATCH (subject:Subject)
	// 	WHERE CASE
	// 		WHEN $subject IS NOT NULL
	// 		THEN subject.name = $subject
	// 		ELSE TRUE
	// 	END

	// 	MATCH (resource:Resource)
	// 	WHERE CASE
	// 		WHEN $resource IS NOT NULL
	// 		THEN resource.name = $resource
	// 		ELSE TRUE
	// 	END

	// 	WITH rel, resource, subject, policy, specifier, count(DISTINCT s.key) AS matches, size(keys(normalizedSpecifiers)) AS requiredMatches
	// 		WHERE matches = requiredMatches

	// 	RETURN DISTINCT policy.id AS policyId, subject, resource, type(rel) AS action, collect(specifier) AS specifiers
	// `

	// query := `
	// // 1. We start immediately with a Subquery using UNION
	// CALL () {
	// 	// --- BRANCH A: Specifiers Input is NULL ---
	// 	// If null, we simply match ANY policy with the correct action
	// 	WITH $specifiers AS inputMap
	// 	WHERE inputMap IS NULL

	// 	MATCH (policy:Policy)-[rel]->(specifier:Specifier)
	// 	WHERE type(rel) = $action

	// 	RETURN policy, rel, specifier

	// 	UNION

	// 	// --- BRANCH B: Specifiers Input HAS DATA ---
	// 	// If not null, we perform the complex intersection check
	// 	WITH $specifiers AS inputMap
	// 	WHERE inputMap IS NOT NULL

	// 	// B1. Fetch all DB keys to handle implicit wildcards
	// 	CALL () {
	// 		MATCH (s:Specifier) WHERE s.key <> "*"
	// 		RETURN collect(DISTINCT s.key) AS allDbKeys
	// 	}

	// 	// B2. Normalize the map (Fill missing keys with '*')
	// 	WITH inputMap,
	// 		reduce(acc = inputMap, k IN allDbKeys |
	// 			CASE
	// 				WHEN k IN keys(inputMap) THEN acc
	// 				ELSE apoc.map.setKey(acc, k, "*")
	// 			END
	// 		) AS normalizedSpecifiers

	// 	// B3. Unwind and Match
	// 	UNWIND keys(normalizedSpecifiers) AS k
	// 	WITH normalizedSpecifiers, k, normalizedSpecifiers[k] AS v

	// 	MATCH (s:Specifier)
	// 	WHERE s.key = k AND s.value = v

	// 	MATCH (policy:Policy)-[rel]->(s)
	// 	WHERE type(rel) = $action

	// 	// B4. Aggregation & Verification
	// 	// We collect the specifiers to keep them associated with the policy
	// 	WITH policy, rel, collect(s) AS policySpecifiers,
	// 		count(DISTINCT s.key) AS matches,
	// 		size(keys(normalizedSpecifiers)) AS requiredMatches

	// 	// B5. Filter: The policy must match EVERY key in the normalized map
	// 	WHERE matches = requiredMatches

	// 	WITH DISTINCT policy, rel, policySpecifiers

	// 	// B6. Unwind back to rows to match the format of Branch A
	// 	UNWIND policySpecifiers AS specifier
	// 	RETURN DISTINCT policy, rel, specifier
	// }

	// // 2. Global Filters for Subject and Resource
	// // These apply to the results of EITHER branch above
	// MATCH (subject:Subject)
	// WHERE ($subject IS NULL OR subject.name = $subject)

	// MATCH (resource:Resource)
	// WHERE ($resource IS NULL OR resource.name = $resource)

	// // 3. Final Aggregation
	// RETURN DISTINCT
	// 	policy.id AS policyId,
	// 	subject,
	// 	resource,
	// 	type(rel) AS action,
	// 	collect(specifier) AS specifiers
	// `

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
