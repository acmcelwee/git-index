package cmd

import (
	"fmt"
	"os"

	"github.com/acmcelwee/git-index/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-index",
	Short: "A blazingly fast git repository jumper",
	Long:  `git-index allows you to quickly jump between scattered git repositories.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if err := config.InitConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing config:", err)
		os.Exit(1)
	}
}
