package main

import (
	"github.com/namsnath/otter/cmd"
	"github.com/namsnath/otter/db"
)

func main() {
	db.SetupInstance("bolt://localhost:7687", "neo4j", "password")
	cmd.Execute()
}
