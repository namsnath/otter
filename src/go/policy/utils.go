package policy

import (
	"errors"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ErrPolicyIDRequired = errors.New("policy ID is required")

func ProcessPolicyRecord(record *neo4j.Record) (Policy, error) {
	policy := Policy{}

	policyIdVal, _ := record.Get("policyId")
	policy.Id = policyIdVal.(string)

	actionVal, _ := record.Get("action")
	actionStr := actionVal.(string)
	actionEnum, err := action.FromString(actionStr)
	if err != nil {
		return Policy{}, err
	}
	policy.Action = actionEnum

	subjectVal, _ := record.Get("subject")
	subjectNode := subjectVal.(neo4j.Node)
	subjectType, err := subject.SubjectTypeFromString(subjectNode.Props["type"].(string))
	if err != nil {
		return Policy{}, err
	}
	subjectObj := subject.Subject{Name: subjectNode.Props["name"].(string), Type: subjectType}
	policy.Subject = subjectObj

	resourceVal, _ := record.Get("resource")
	resourceNode := resourceVal.(neo4j.Node)
	resourceObj := resource.Resource{Name: resourceNode.Props["name"].(string)}
	policy.Resource = resourceObj

	specifiersVal, _ := record.Get("specifiers")
	specifiers := specifier.SpecifierGroup{Specifiers: []specifier.Specifier{}}
	if specifiersVal != nil {
		specifierNodes := specifiersVal.([]interface{})
		for _, specifierNode := range specifierNodes {
			specifierNode := specifierNode.(neo4j.Node)
			key := specifierNode.Props["key"].(string)
			value := specifierNode.Props["value"].(string)
			specifiers.Specifiers = append(specifiers.Specifiers, specifier.Specifier{Key: key, Value: value})
		}
	}
	policy.Specifiers = specifiers

	return policy, nil
}
