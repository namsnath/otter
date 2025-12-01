package action

import "fmt"

type Action string

const (
	ActionRead  Action = "READ"
	ActionWrite Action = "WRITE"
)

var ErrInvalidAction = fmt.Errorf("invalid Action")

func FromString(s string) (Action, error) {
	switch s {
	case "READ":
		return ActionRead, nil
	case "WRITE":
		return ActionWrite, nil
	default:
		return "", ErrInvalidAction
	}
}
