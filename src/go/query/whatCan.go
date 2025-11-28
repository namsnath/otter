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

	// TODO: Check if this honors all the specifiers specified in the policy
	// Policy will say {Role: "admin", Env: "prod"}.
	// If query is only {Role: "admin"}, should it return resources?
	query := `
		// Unwind specifiers to match each one
		UNWIND keys($specifiers) AS k

		MATCH (subject:Subject {name: $subject})
		MATCH (specifier:Specifier {key: k, value: $specifiers[k]})
		MATCH (parent:Resource {name: $parent})

		// All policies applicable to the subject (directly or inherited)
		MATCH (subject)-[:CHILD_OF*0..]->(:Subject)-[:HAS_POLICY]->(policy:Policy)

		// Policies that have the required action and specifier
		MATCH (policy)-[rel:$($action)]->(s:Specifier)<-[:CHILD_OF*0..]-(specifier)

		// Resources that apply to the policies
		MATCH (policy)<-[:HAS_POLICY]-(:Resource)<-[:CHILD_OF*0..]-(resource:Resource)

		// Resources under the specified parent resource
		MATCH (parent)<-[:CHILD_OF*0..]-(resource:Resource)

		RETURN DISTINCT resource.name AS resource
	`
	params := map[string]any{
		"subject":    qb.subject.Name,
		"action":     string(qb.action),
		"parent":     qb.underResource.Name,
		"specifiers": qb.specifiers,
	}

	if len(qb.specifiers) == 0 {
		params["specifiers"] = map[string]string{"*": "*"}
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
		"specifier", "*=*",
		"resources", resources,
		"duration", result.Summary.ResultAvailableAfter(),
	)

	return resources, nil
}
