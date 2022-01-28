package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gozip",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Simple zip tool v0.0.1")
		return nil
	},
}
