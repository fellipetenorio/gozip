package cmd

import (
	"github.com/spf13/cobra"
)

var Verbose bool

var rootCmd = &cobra.Command{
	Use:   "gozip",
	Short: "Short Descriptoion",
	Long:  "Hey lon gshort",
}

func Execute() error {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	return rootCmd.Execute()
}

func init() {
	//cobra.OnInitialize(initConfig)

	//rootCmd.AddCommand(listCmd)
}
