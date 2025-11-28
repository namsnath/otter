package specifier

import (
	"fmt"

	"github.com/namsnath/otter/db"
)

func (s Specifier) Create() Specifier {
	db.ExecuteQuery(
		"CREATE (r:Specifier {key: $key, value: $value})",
		map[string]any{
			"key":   s.Key,
			"value": s.Value,
		},
	)

	return s
}

func (s Specifier) CreateAsChildOf(parent Specifier) (Specifier, error) {
	if s.Key != parent.Key && parent.Key != "*" {
		return Specifier{}, fmt.Errorf("cannot create child specifier with different key except under `*`: %s vs %s", s.Key, parent.Key)
	}
	if parent.Key == "*" && s.Key == "*" {
		return Specifier{}, fmt.Errorf("cannot create child specifier with key `*` under another `*`. This is a special root node")
	}

	db.ExecuteQuery(`
		CREATE (s:Specifier {key: $key, value: $value})
		WITH s
		MATCH (p:Specifier {key: $parentKey, value: $parentValue})
		CREATE (s)-[:CHILD_OF]->(p)
		`,
		map[string]any{
			"key":         s.Key,
			"value":       s.Value,
			"parentKey":   parent.Key,
			"parentValue": parent.Value,
		},
	)

	return s, nil
}
