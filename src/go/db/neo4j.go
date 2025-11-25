package db

import (
	"context"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	instance *Neo4J
	once     sync.Once
)

type Neo4J struct {
	ctx    context.Context
	driver neo4j.DriverWithContext
}

func GetInstance() *Neo4J {
	once.Do(func() {
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

		err = driver.VerifyConnectivity(ctx)
		if err != nil {
			panic(err)
		}

		instance = &Neo4J{
			ctx:    ctx,
			driver: driver,
		}
	})

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
