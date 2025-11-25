package subject

import "fmt"

type SubjectType string

const (
	SubjectTypePrincipal SubjectType = "Principal"
	SubjectTypeGroup     SubjectType = "Group"
)

func SubjectTypeFromString(s string) SubjectType {
	switch s {
	case "Principal":
		return SubjectTypePrincipal
	case "Group":
		return SubjectTypeGroup
	default:
		panic(fmt.Sprintf("Invalid SubjectType %v", s))
	}
}

type Subject struct {
	Name string
	Type SubjectType
}
