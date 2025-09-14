package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Runs the CSI controller plugin",
	Long:  `Runs the CSI controller plugin.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: run the controller plugin")
	},
}

func init() {
	rootCmd.AddCommand(controllerCmd)
}
