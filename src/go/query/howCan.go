package query

import (
	"fmt"
	"log/slog"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
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

func (qb HowCanQueryBuilder) Query() ([]map[string]string, error) {
	// TODO: Support specifiers. Some specifiers can be provided, we return the rest (excluding keys in the input).
	// TODO: If a root specifier is accessible, don't expand to children. See if this can be done in the query itself.

	qb, validationError := qb.Validate()
	if validationError != nil {
		return []map[string]string{}, validationError
	}

	query := `
		MATCH (subject:Subject {name: $subject, type: $subjectType})
		MATCH (resource:Resource {name: $resource})
		MATCH (subject)-[:CHILD_OF*0..]->(:Subject)-[:HAS_POLICY]->(policy:Policy)
		MATCH (resource)-[:CHILD_OF*0..]->(:Resource)-[:HAS_POLICY]->(policy:Policy)

		WITH policy
			MATCH (policy)-[rel:$($action)]->(:Specifier)<-[:CHILD_OF*0..]-(specifier:Specifier)

		RETURN DISTINCT policy.id as policyId, specifier.key AS specifierKey, collect(specifier.value) as specifierVals
	`

	params := map[string]any{
		"subject":     qb.subject.Name,
		"subjectType": qb.subject.Type,
		"action":      qb.action,
		"resource":    qb.resource.Name,
	}

	result := db.ExecuteQuery(query, params)

	specifierMaps := []map[string]string{}
	policyMap := map[string]map[string][]string{}

	for _, record := range result.Records {
		policyIdVal, policyIdOk := record.Get("policyId")
		specifierKeyVal, specifierKeyOk := record.Get("specifierKey")
		specifierValsVal, specifierValsOk := record.Get("specifierVals")

		if !policyIdOk || !specifierKeyOk || !specifierValsOk {
			return []map[string]string{}, fmt.Errorf("unexpected result format from HowCan query")
		}

		policyStr, policyStrOk := policyIdVal.(string)
		specifierKey, specifierKeyOk := specifierKeyVal.(string)
		specifierVals, specifierValsOk := specifierValsVal.([]any)

		if !policyStrOk || !specifierKeyOk || !specifierValsOk {
			return []map[string]string{}, fmt.Errorf("unexpected result types from HowCan query")
		}

		if _, exists := policyMap[policyStr]; !exists {
			policyMap[policyStr] = map[string][]string{}
		}
		policyMap[policyStr][specifierKey] = []string{}

		for _, val := range specifierVals {
			if valStr, valStrOk := val.(string); valStrOk {
				policyMap[policyStr][specifierKey] = append(policyMap[policyStr][specifierKey], valStr)
			} else {
				return []map[string]string{}, fmt.Errorf("unexpected specifier value type from HowCan query")
			}
		}
	}

	// TODO: Cross product of all specifiers in this policy
	slog.Info("Specifiers", "policyMap", policyMap)

	return specifierMaps, nil
}
