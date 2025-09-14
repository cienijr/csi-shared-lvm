package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Runs the CSI node plugin",
	Long:  `Runs the CSI node plugin.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: run the node plugin")
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
