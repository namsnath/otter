package query

import (
	"log/slog"

	"github.com/namsnath/gatekeeper/action"
	"github.com/namsnath/gatekeeper/db"
	"github.com/namsnath/gatekeeper/resource"
	"github.com/namsnath/gatekeeper/specifier"
	"github.com/namsnath/gatekeeper/subject"
)

type WhatCanQueryBuilder struct {
	subject    subject.Subject
	action     action.Action
	resource   *resource.Resource
	specifiers map[string]string
}

func WhatCan(subject subject.Subject) *WhatCanQueryBuilder {
	return &WhatCanQueryBuilder{subject: subject}
}

func (qb *WhatCanQueryBuilder) Perform(action action.Action) *WhatCanQueryBuilder {
	qb.action = action
	return qb
}

func (qb *WhatCanQueryBuilder) Under(resource *resource.Resource) *WhatCanQueryBuilder {
	qb.resource = resource
	return qb
}

func (qb *WhatCanQueryBuilder) With(specifiers *specifier.SpecifierGroup) *WhatCanQueryBuilder {
	qb.specifiers = specifiers.AsMap()
	return qb
}

func (qb *WhatCanQueryBuilder) Execute() []resource.Resource {
	// TODO: Improve the specifiers conditional to check for supersets of the defined specifiers
	query := `
		MATCH (s:Subject {name: $subjectName})-[:CHILD_OF*0..]->(:Subject)-[rel:$($action)]->(:Resource)<-[:CHILD_OF*0..]-(r:Resource)
		WHERE properties(rel) = $specifiers
		RETURN DISTINCT r.name AS resource
	`
	params := map[string]any{
		"subjectName": qb.subject.Name,
		"action":      string(qb.action),
		"specifiers":  qb.specifiers,
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

	slog.Info("WhatCan",
		"subject", qb.subject,
		"action", qb.action,
		"resource", qb.resource,
		"specifiers", qb.specifiers,
		"resources", resources,
		"duration", result.Summary.ResultAvailableAfter())

	return resources
}
