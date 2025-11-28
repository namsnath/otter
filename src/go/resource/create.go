package resource

import "github.com/namsnath/otter/db"

func (resource Resource) Create() Resource {
	db.ExecuteQuery(`
		CREATE (r:Resource {name: $name})
		`,
		map[string]any{
			"name": resource.Name,
		},
	)

	return resource
}

func (resource Resource) CreateAsChildOf(parent Resource) Resource {
	db.ExecuteQuery(`
		CREATE (r:Resource {name: $name})
		WITH r
		MATCH (p:Resource {name: $parentName})
		CREATE (r)-[:CHILD_OF]->(p)
		`,
		map[string]any{
			"name":       resource.Name,
			"parentName": parent.Name,
		},
	)

	return resource
}
