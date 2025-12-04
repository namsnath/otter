package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	tcNeo4j "github.com/testcontainers/testcontainers-go/modules/neo4j"
)

var (
	instance *Neo4J
	once     sync.Once
)

type Neo4J struct {
	ctx    context.Context
	driver neo4j.DriverWithContext
}

func SetupInstance(dbUri, dbUser, dbPassword string) {
	once.Do(func() {
		ctx := context.Background()

		driver, err := neo4j.NewDriverWithContext(
			dbUri,
			neo4j.BasicAuth(dbUser, dbPassword, ""))

		if err != nil {
			panic(err)
		}

		err = driver.VerifyConnectivity(ctx)
		if err != nil {
			panic(err)
		}

		instance = &Neo4J{
			ctx:    ctx,
			driver: driver,
		}
	})
}

func GetInstance() *Neo4J {
	if instance == nil {
		panic("Neo4J instance not initialized. Call SetupInstance first.")
	}

	return instance
}

func (s *Neo4J) Close() error {
	err := s.driver.Close(s.ctx)
	if err != nil {
		return err
	}
	instance = nil
	return nil
}

func TestContainer() (context.Context, *tcNeo4j.Neo4jContainer) {
	ctx := context.Background()
	testPassword := "password"

	container, err := tcNeo4j.Run(
		ctx,
		"neo4j:latest",
		tcNeo4j.WithLabsPlugin(tcNeo4j.Apoc),
		tcNeo4j.WithAdminPassword(testPassword),
	)
	if err != nil {
		panic(err)
	}

	container.Start(ctx)

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	mappedPort, err := container.MappedPort(ctx, "7687/tcp") // Default Bolt port
	if err != nil {
		panic(err)
	}

	SetupInstance(fmt.Sprintf("bolt://%s:%d", host, mappedPort.Int()), "neo4j", testPassword)

	return ctx, container
}
