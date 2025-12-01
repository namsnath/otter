package cmd

import (
	"github.com/namsnath/otter/db"
	"github.com/namsnath/otter/query"
	"github.com/spf13/cobra"
)

var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Delete everything and set up test state in the database",
	Run: func(cmd *cobra.Command, args []string) {
		instance := db.GetInstance()
		query.DeleteEverything()
		query.SetupTestState()
		instance.Close()
	},
}

func init() {}
