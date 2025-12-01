package cmd

import (
	"os"

	query "github.com/namsnath/otter/cmd/query"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "otter",
	Short: "otter: graph-based authorization system",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(query.QueryCmd)
	RootCmd.AddCommand(SetupCmd)
}
