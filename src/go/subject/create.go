package subject

import "github.com/namsnath/otter/db"

func (subject Subject) Create() Subject {
	db.ExecuteQuery(`
		CREATE (s:Subject {name: $name, type: $type})
		`,
		map[string]any{
			"name": subject.Name,
			"type": subject.Type,
		},
	)

	return subject
}

func (subject Subject) CreateAsChildOf(parent Subject) Subject {
	db.ExecuteQuery(`
		CREATE (s:Subject {name: $name, type: $type})
		WITH s
		MATCH (p:Subject {name: $parentName, type: $parentType})
		CREATE (s)-[:CHILD_OF]->(p)
		`,
		map[string]any{
			"name":       subject.Name,
			"type":       subject.Type,
			"parentName": parent.Name,
			"parentType": parent.Type,
		},
	)

	return subject
}
