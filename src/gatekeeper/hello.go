package main

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/nivohavi/go-cypher-dsl/pkg/cypher"
)

func ExecuteQuery(ctx context.Context, driver neo4j.DriverWithContext, query string, params map[string]any) *neo4j.EagerResult {
	result, err := neo4j.ExecuteQuery(ctx, driver, query, params,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		panic(err)
	}
	return result
}

// DeleteEverything deletes all nodes and relationships in the database
func DeleteEverything(ctx context.Context, driver neo4j.DriverWithContext) {
	result := ExecuteQuery(ctx, driver, `
		MATCH (n) DETACH DELETE n
		`,
		nil)
	fmt.Printf("All nodes and relationships deleted in %+v\n", result.Summary.ResultAvailableAfter())
}

func setupBasicNodes(ctx context.Context, driver neo4j.DriverWithContext) {
	// ExecuteQuery(ctx, driver, `CREATE CONSTRAINT resource_name_constraint IF NOT EXISTS FOR (r:Resource) REQUIRE r.name IS UNIQUE`, nil)
	ExecuteQuery(ctx, driver, `CREATE INDEX subject_name_index IF NOT EXISTS FOR (s:Subject) ON (s.name)`, nil)
	ExecuteQuery(ctx, driver, `CREATE INDEX resource_name_index IF NOT EXISTS FOR (r:Resource) ON (r.name)`, nil)

	result := ExecuteQuery(ctx, driver, `
		CREATE (p1:Subject {type: 'Principal', name: "Principal1"})
		CREATE (p2:Subject {type: 'Principal', name: "Principal2"})
		CREATE (p3:Subject {type: 'Principal', name: "Principal3"})
		CREATE (g1:Subject {type: 'Group', name: "Group1"})
		CREATE (g2:Subject {type: 'Group', name: "Group2"})
		CREATE (p1)-[:CHILD_OF]->(g1)
		CREATE (g1)-[:CHILD_OF]->(g2)
		CREATE (p2)-[:CHILD_OF]->(g2)

		CREATE (rRoot:Resource {type: 'Resource', name: '_'})
		CREATE (r1:Resource {type: 'Resource', name: "Resource1"})
		CREATE (r2:Resource {type: 'Resource', name: "Resource2"})
		CREATE (r1)-[:CHILD_OF]->(rRoot)
		CREATE (r2)-[:CHILD_OF]->(rRoot)
		CREATE (g1)-[:READ]->(r1)
		CREATE (g2)-[:READ]->(r2)
		CREATE (p2)-[:READ]->(r1)
		CREATE (p3)-[:READ {specifier: "someSpecifier"}]->(rRoot)
		`, nil)

	summary := result.Summary
	fmt.Printf("Created %v nodes and %v relationships in %+v.\n",
		summary.Counters().NodesCreated(),
		summary.Counters().RelationshipsCreated(),
		summary.ResultAvailableAfter())
}

func SubjectCanDo(ctx context.Context, driver neo4j.DriverWithContext, subject any, action string, resource any) bool {
	cypher.Node("Subject").Named("s")
	result := ExecuteQuery(ctx, driver, `
		MATCH (s:Subject {name: $subjectName})
		MATCH (r:Resource {name: $resourceName})
		RETURN EXISTS { (s)-[:CHILD_OF*0..]->(:Subject)-[rel:READ]->(:Resource)<-[:CHILD_OF*0..]-(r) } as CanDo
		`,
		map[string]any{"subjectName": subject, "resourceName": resource})

	if len(result.Records) == 0 {
		return false
	}

	val, ok := result.Records[0].Get("CanDo")
	if !ok {
		return false
	}

	fmt.Printf("SubjectCanDo '%s': %v in %+v\n", resource, val, result.Summary.ResultAvailableAfter())

	return val.(bool)
}

func SubjectsThatCanDo(ctx context.Context, driver neo4j.DriverWithContext, resourceName string) []string {
	// Fetch subjects with transitive CHILD_OF relationships that can access the given resource
	// Second part of the query gets all the resources that could be parents of the given resource
	// First part then gets all subjects that can READ any of those resources
	// Subject s is a child of any Subject that can READ a Resource that is a parent of any Resouce r, where r.name = $resourceName
	result := ExecuteQuery(ctx, driver, `
		MATCH (s:Subject)-[:CHILD_OF*0..]->()-[:READ]->()<-[:CHILD_OF*0..]-(r:Resource {name: $resourceName})
		RETURN DISTINCT s.name AS subject
		`,
		map[string]any{"resourceName": resourceName})

	subjects := make([]string, 0, len(result.Records))
	for _, record := range result.Records {
		if name, ok := record.Get("subject"); ok {
			if str, ok := name.(string); ok {
				subjects = append(subjects, str)
			}
		}
	}

	fmt.Printf("SubjectsThatCanDo '%s': %v in %+v\n", resourceName, subjects, result.Summary.ResultAvailableAfter())
	return subjects
}

func main() {
	ctx := context.Background()
	dbUri := "bolt://localhost:7687"
	dbUser := "neo4j"
	dbPassword := "password"
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		panic(err)
	}
	defer driver.Close(ctx)

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established.")
	DeleteEverything(ctx, driver)
	setupBasicNodes(ctx, driver)
	SubjectCanDo(ctx, driver, "Principal1", "READ", "Resource1")
	SubjectCanDo(ctx, driver, "Group2", "READ", "Resource2")
	SubjectCanDo(ctx, driver, "Principal3", "READ", "_")
	SubjectsThatCanDo(ctx, driver, "Resource1")
	SubjectsThatCanDo(ctx, driver, "Resource2")
	SubjectsThatCanDo(ctx, driver, "_")
}
