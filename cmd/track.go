package cmd

import (
	"fmt"
	"os"

	"github.com/acmcelwee/git-index/internal/frecency"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:    "track [path]",
	Short:  "Track repository access for frecency ranking",
	Args:   cobra.ExactArgs(1),
	Hidden: true, // Keep it hidden from typical help output
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		if err := frecency.Track(path); err != nil {
			fmt.Fprintln(os.Stderr, "Error tracking path:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
}
