package subject

import "fmt"

type SubjectType string

const (
	SubjectTypePrincipal SubjectType = "Principal"
	SubjectTypeGroup     SubjectType = "Group"
)

var ErrInvalidSubjectType = fmt.Errorf("invalid SubjectType")

func SubjectTypeFromString(s string) (SubjectType, error) {
	switch s {
	case "Principal":
		return SubjectTypePrincipal, nil
	case "Group":
		return SubjectTypeGroup, nil
	default:
		return "", ErrInvalidSubjectType
	}
}

type Subject struct {
	Name string
	Type SubjectType
}
