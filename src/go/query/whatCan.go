package query

import (
	"log/slog"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
)

type WhatCanQueryBuilder struct {
	subject       subject.Subject
	action        action.Action
	underResource resource.Resource
	specifiers    map[string]string
}

func WhatCan(subject subject.Subject) WhatCanQueryBuilder {
	return WhatCanQueryBuilder{subject: subject}
}

func (qb WhatCanQueryBuilder) Perform(action action.Action) WhatCanQueryBuilder {
	qb.action = action
	return qb
}

func (qb WhatCanQueryBuilder) Under(underResource resource.Resource) WhatCanQueryBuilder {
	qb.underResource = underResource
	return qb
}

func (qb WhatCanQueryBuilder) With(specifiers specifier.SpecifierGroup) WhatCanQueryBuilder {
	qb.specifiers = specifiers.AsMap()
	return qb
}

func (qb WhatCanQueryBuilder) Query() ([]resource.Resource, error) {
	if qb.underResource.Name == "" {
		qb.underResource = resource.Resource{Name: "_"}
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
		"parent":     qb.underResource.Name,
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
		"underResource", qb.underResource,
		"specifiers", qb.specifiers,
		"resources", resources,
		"duration", result.Summary.ResultAvailableAfter(),
	)

	return resources, nil
}
