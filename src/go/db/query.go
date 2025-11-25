package db

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ExecuteQuery(query string, params map[string]any) *neo4j.EagerResult {
	instance := GetInstance()
	result, err := neo4j.ExecuteQuery(instance.ctx, instance.driver, query, params,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		panic(err)
	}
	return result
}
