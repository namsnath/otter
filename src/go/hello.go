package main

import (
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/query"
)

func main() {
	instance := db.GetInstance()
	query.DeleteEverything()
	query.SetupTestState()
	instance.Close()
}
