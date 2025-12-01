package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var QueryCmd = &cobra.Command{
	Use:   "query",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {}
