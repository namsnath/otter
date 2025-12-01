package query

import (
	"errors"
	"log/slog"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
	"github.com/namsnath/otter/utils/hashset"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type WhatCanQueryBuilder struct {
	subject        subject.Subject
	action         action.Action
	parentResource resource.Resource
	specifiers     map[string]string
}

var ErrSubjectNotSet = errors.New("subject not set in query builder")
var ErrActionNotSet = errors.New("action not set in query builder")
var ErrParentResourceNotSet = errors.New("parentResource not set in query builder")

func WhatCan(subject subject.Subject) WhatCanQueryBuilder {
	return WhatCanQueryBuilder{subject: subject}
}

func (qb WhatCanQueryBuilder) Perform(action action.Action) WhatCanQueryBuilder {
	qb.action = action
	return qb
}

func (qb WhatCanQueryBuilder) Under(parentResource resource.Resource) WhatCanQueryBuilder {
	qb.parentResource = parentResource
	return qb
}

func (qb WhatCanQueryBuilder) With(specifiers specifier.SpecifierGroup) WhatCanQueryBuilder {
	qb.specifiers = specifiers.AsMap()
	return qb
}

func (qb WhatCanQueryBuilder) Validate() (WhatCanQueryBuilder, error) {
	if qb.subject == (subject.Subject{}) {
		return qb, ErrSubjectNotSet
	}

	if qb.action == "" {
		return qb, ErrActionNotSet
	}

	if qb.parentResource == (resource.Resource{}) {
		return qb, ErrParentResourceNotSet
	}

	return qb, nil
}

func (qb WhatCanQueryBuilder) Query() ([]resource.Resource, error) {
	qb, err := qb.Validate()
	if err != nil {
		return nil, err
	}

	query := `
		MATCH (specifier:Specifier)
		WHERE specifier.key <> "*"
		WITH collect(DISTINCT specifier.key) AS allKeys
		WITH reduce(specMap = $specifiers, k IN allKeys |
			CASE WHEN NOT k IN keys(specMap) THEN apoc.map.setKey(specMap, k, "*") ELSE specMap END
		) AS normalizedSpecifiers

		UNWIND keys(normalizedSpecifiers) AS k
		WITH normalizedSpecifiers, k, normalizedSpecifiers[k] AS v

		MATCH (s:Specifier)
		WHERE s.key = k AND s.value = v

		MATCH (p:Policy)-[:$($action)]->(ps:Specifier)<-[:CHILD_OF*0..]-(s)

		MATCH (subject:Subject {name: $subject})-[:CHILD_OF*0..]->(parents:Subject)-[:HAS_POLICY]->(p)

		WITH p, count(DISTINCT s.key) AS matches, size(keys(normalizedSpecifiers)) AS requiredMatches, normalizedSpecifiers
		WHERE matches = requiredMatches

		MATCH (resource:Resource)-[:CHILD_OF*0..]->(:Resource)-[:HAS_POLICY]->(p)
		MATCH (resource)-[:CHILD_OF*0..]->(parent:Resource {name: $parent})

		RETURN DISTINCT resource.name AS resource
	`

	params := map[string]any{
		"subject":    qb.subject.Name,
		"action":     string(qb.action),
		"parent":     qb.parentResource.Name,
		"specifiers": qb.specifiers,
	}

	result := db.ExecuteQuery(query, params)

	resources := make([]resource.Resource, 0, len(result.Records))
	for _, record := range result.Records {
		nameVal, nameOk := record.Get("resource")
		if nameOk {
			if nameStr, nameIsStr := nameVal.(string); nameIsStr {
				resource := resource.Resource{Name: nameStr}
				resources = append(resources, resource)
			}
		}
	}

	slog.Info(
		"WhatCan",
		"subject", qb.subject,
		"action", qb.action,
		"underResource", qb.parentResource,
		"specifiers", qb.specifiers,
		"resources", resources,
		"duration", result.Summary.ResultAvailableAfter(),
		"rows", len(result.Records),
	)

	return resources, nil
}

// Retrieve resources for a given subject, action, specifiers, and a parent resource, expanding to fetch all additional specifiers
//
// Returns:
//   - A mapping of resources to their specifiers
//   - error
func (qb WhatCanQueryBuilder) QueryWithoutAllSpecifiers() (map[resource.Resource]map[string][]specifier.Specifier, error) {
	qb, err := qb.Validate()
	if err != nil {
		return nil, err
	}

	query := `
		WITH $specifiers AS normalizedSpecifiers
			UNWIND keys(normalizedSpecifiers) AS k

		WITH normalizedSpecifiers, k, normalizedSpecifiers[k] AS v
			MATCH (s:Specifier)
				WHERE s.key = k AND s.value = v

			MATCH (p:Policy)-[:$($action)]->(:Specifier)<-[:CHILD_OF*0..]-(s)

			MATCH (subject:Subject {name: $subject})-[:CHILD_OF*0..]->(:Subject)-[:HAS_POLICY]->(p)

		WITH p, count(DISTINCT s.key) AS matches, size(keys(normalizedSpecifiers)) AS requiredMatches, normalizedSpecifiers
			WHERE matches = requiredMatches

			MATCH (resource:Resource)-[:CHILD_OF*0..]->(:Resource)-[:HAS_POLICY]->(p)
			MATCH (resource)-[:CHILD_OF*0..]->(parent:Resource {name: $parent})

			// Expand the graph to get all the specifiers
			MATCH (p)-[:$($action)]->(parentSpecifier:Specifier)<-[:CHILD_OF*0..]-(otherSpecifier:Specifier)
				WHERE NOT otherSpecifier.key IN keys(normalizedSpecifiers) AND parentSpecifier.key <> "*"

		RETURN resource, otherSpecifier
	`

	params := map[string]any{
		"subject":    qb.subject.Name,
		"action":     string(qb.action),
		"parent":     qb.parentResource.Name,
		"specifiers": qb.specifiers,
	}

	result := db.ExecuteQuery(query, params)

	resourcesWithSpecifiersMap := map[resource.Resource]map[string]*hashset.HashSet[string]{}
	for _, record := range result.Records {
		recordMap := record.AsMap()
		resourceVal := recordMap["resource"].(neo4j.Node).Props
		specifierVal := recordMap["otherSpecifier"].(neo4j.Node).Props

		resourceObj := resource.Resource{Name: resourceVal["name"].(string)}
		specifierObj := specifier.Specifier{Key: specifierVal["key"].(string), Value: specifierVal["value"].(string)}

		if _, exists := resourcesWithSpecifiersMap[resourceObj]; !exists {
			resourcesWithSpecifiersMap[resourceObj] = map[string]*hashset.HashSet[string]{}
		}
		if _, exists := resourcesWithSpecifiersMap[resourceObj][specifierObj.Key]; !exists {
			resourcesWithSpecifiersMap[resourceObj][specifierObj.Key] = hashset.New[string]()
		}

		resourcesWithSpecifiersMap[resourceObj][specifierObj.Key].Add(specifierObj.Value)
	}

	resourcesWithSpecifiers := map[resource.Resource]map[string][]specifier.Specifier{}
	for res, specMap := range resourcesWithSpecifiersMap {
		resourcesWithSpecifiers[res] = map[string][]specifier.Specifier{}
		for specKey, specValues := range specMap {
			specList := make([]specifier.Specifier, 0, specValues.Len())
			for specValue := range specValues.All() {
				specList = append(specList, specifier.Specifier{Key: specKey, Value: specValue})
			}
			resourcesWithSpecifiers[res][specKey] = specList
		}
	}

	slog.Info(
		"WhatCan QueryWithoutAllSpecifiers",
		"subject", qb.subject,
		"action", qb.action,
		"underResource", qb.parentResource,
		"specifiers", qb.specifiers,
		"resourcesWithSpecifiers", resourcesWithSpecifiers,
		"duration", result.Summary.ResultAvailableAfter(),
		"rows", len(result.Records),
	)

	return resourcesWithSpecifiers, nil
}
