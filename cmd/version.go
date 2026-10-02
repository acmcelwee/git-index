package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of git-index",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("git-index version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
