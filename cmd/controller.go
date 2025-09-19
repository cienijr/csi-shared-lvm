package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/cienijr/csi-shared-lvm/pkg/driver"
	"github.com/cienijr/csi-shared-lvm/pkg/server"
)

var (
	controllerEndpoint string
)

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Runs the CSI controller plugin",
	Long:  `Runs the CSI controller plugin.`,
	Run: func(cmd *cobra.Command, args []string) {
		d := driver.NewDriver(controllerEndpoint)
		s := server.New(d, d, nil)
		if err := s.Run(controllerEndpoint); err != nil {
			klog.Fatalf("error running server: %v", err)
		}
	},
}

func init() {
	controllerCmd.PersistentFlags().StringVar(&controllerEndpoint, "endpoint", "unix:///tmp/csi.sock", "The endpoint for the CSI driver.")
	rootCmd.AddCommand(controllerCmd)
}
