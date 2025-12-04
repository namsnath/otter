package query

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
	"github.com/namsnath/otter/utils"
	"github.com/namsnath/otter/utils/hashset"
)

type HowCanQueryBuilder struct {
	subject    subject.Subject
	action     action.Action
	resource   resource.Resource
	specifiers map[string]string
}

func HowCan(subject subject.Subject) HowCanQueryBuilder {
	return HowCanQueryBuilder{subject: subject}
}

func (qb HowCanQueryBuilder) Perform(action action.Action) HowCanQueryBuilder {
	qb.action = action
	return qb
}

func (qb HowCanQueryBuilder) On(resource resource.Resource) HowCanQueryBuilder {
	qb.resource = resource
	return qb
}

func (qb HowCanQueryBuilder) With(specifiers specifier.SpecifierGroup) HowCanQueryBuilder {
	qb.specifiers = specifiers.AsMap()
	return qb
}

func (qb HowCanQueryBuilder) Validate() (HowCanQueryBuilder, error) {
	if qb.subject == (subject.Subject{}) || qb.action == "" || qb.resource == (resource.Resource{}) {
		return qb, fmt.Errorf("incomplete HowCan query: subject, action, and resource must be set")
	}

	return qb, nil
}

func (qb HowCanQueryBuilder) Query() ([]specifier.SpecifierGroup, error) {
	qb, validationError := qb.Validate()
	if validationError != nil {
		return []specifier.SpecifierGroup{}, validationError
	}

	query := `
		MATCH (s:Subject {name: $subject, type: $subjectType})-[:CHILD_OF*0..]->(sParent)
		MATCH (r:Resource {name: $resource})-[:CHILD_OF*0..]->(rParent)

		MATCH (sParent)-[:HAS_POLICY]->(policy:Policy)<-[:HAS_POLICY]-(rParent)

		// All the root specifiers for this policy, filtered by the required action
		MATCH (policy)-[rel]->(rootSpec:Specifier)
		WHERE type(rel) = $action

		WITH policy, collect(rootSpec) AS policyRootSpecs

		// If specifiers are provided, the policy is only valid if it covers ALL provided keys.
		WHERE $specifiers IS NULL OR ALL(inputKey IN keys($specifiers) WHERE
			ANY(pSpec IN policyRootSpecs WHERE
				pSpec.key = inputKey AND
				// Check: Is the Input Value a valid descendant of the Policy Specifier?
				EXISTS {
					MATCH (pSpec)<-[:CHILD_OF*0..]-(:Specifier {value: $specifiers[inputKey]})
				}
			)
		)

		// Get all specifiers from the policy
		UNWIND policyRootSpecs AS rootSpec

		// Exclude any specifiers that were provided in the input
		WITH policy, rootSpec
		WHERE $specifiers IS NULL OR NOT rootSpec.key IN keys($specifiers)

		// Exapnd each root specifier to its valid children
		CALL {
			// Two WITH statements required to allow filtering on the imported rootSpec

			WITH rootSpec
			// Path A: It is a wildcard root, return only the root
			WITH rootSpec
			WHERE rootSpec.value = '*'
			RETURN rootSpec AS finalSpec

			UNION

			WITH rootSpec
			// Path B: It is specific, return the root and all children
			WITH rootSpec
			WHERE rootSpec.value <> '*'
			MATCH (rootSpec)<-[:CHILD_OF*0..]-(childSpec)
			RETURN childSpec AS finalSpec
		}

		RETURN
			policy.id AS policyId,
			finalSpec.key AS specifierKey,
			collect(DISTINCT finalSpec.value) AS specifierVals
	`

	params := map[string]any{
		"subject":     qb.subject.Name,
		"subjectType": qb.subject.Type,
		"action":      qb.action,
		"resource":    qb.resource.Name,
		"specifiers":  qb.specifiers,
	}

	if len(qb.specifiers) == 0 {
		params["specifiers"] = nil
	}

	result := db.ExecuteQuery(query, params)

	specifierGroups := []specifier.SpecifierGroup{}
	policyMap := map[string]map[string][]string{}

	for _, record := range result.Records {
		policyIdVal, policyIdOk := record.Get("policyId")
		specifierKeyVal, specifierKeyOk := record.Get("specifierKey")
		specifierValsVal, specifierValsOk := record.Get("specifierVals")

		if !policyIdOk || !specifierKeyOk || !specifierValsOk {
			return []specifier.SpecifierGroup{}, fmt.Errorf("unexpected result format from HowCan query")
		}

		policyStr, policyStrOk := policyIdVal.(string)
		specifierKey, specifierKeyOk := specifierKeyVal.(string)
		specifierVals, specifierValsOk := specifierValsVal.([]any)

		if !policyStrOk || !specifierKeyOk || !specifierValsOk {
			return []specifier.SpecifierGroup{}, fmt.Errorf("unexpected result types from HowCan query")
		}

		if _, exists := policyMap[policyStr]; !exists {
			policyMap[policyStr] = map[string][]string{}
		}
		policyMap[policyStr][specifierKey] = []string{}

		for _, val := range specifierVals {
			if valStr, valStrOk := val.(string); valStrOk {
				policyMap[policyStr][specifierKey] = append(policyMap[policyStr][specifierKey], fmt.Sprintf("%s=%s", specifierKey, valStr))
			} else {
				return []specifier.SpecifierGroup{}, fmt.Errorf("unexpected specifier value type from HowCan query")
			}
		}
	}

	// Create unique cartesian product of all specifier values for each policy
	uniqueGroups := hashset.New[string]()
	for _, specMap := range policyMap {
		specifierLists := [][]string{}

		// Sort keys to ensure consistent ordering of lists
		keys := make([]string, 0, len(specMap))
		for k := range specMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			specifierLists = append(specifierLists, specMap[k])
		}

		combinations := utils.CartesianProduct(specifierLists)
		for _, combination := range combinations {
			uniqueGroups.Add(strings.Join(combination, ","))
		}
	}

	// Convert unique group strings to SpecifierGroups
	for groupStr := range uniqueGroups.All() {
		specifierGroup := specifier.SpecifierGroup{}
		specifierPairs := strings.Split(groupStr, ",")
		for _, pair := range specifierPairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				specifierGroup.Specifiers = append(specifierGroup.Specifiers, specifier.Specifier{Key: kv[0], Value: kv[1]})
			}
		}
		specifierGroups = append(specifierGroups, specifierGroup)
	}

	slog.Info(
		"HowCan",
		"subject", qb.subject,
		"action", qb.action,
		"resource", qb.resource,
		"specifiers", qb.specifiers,
		"specifierGroups", specifierGroups,
		"duration", result.Summary.ResultAvailableAfter(),
		"rows", len(result.Records),
	)

	return specifierGroups, nil
}
