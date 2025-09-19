package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/cienijr/csi-shared-lvm/pkg/driver"
	"github.com/cienijr/csi-shared-lvm/pkg/server"
)

var (
	nodeEndpoint string
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Runs the CSI node plugin",
	Long:  `Runs the CSI node plugin.`,
	Run: func(cmd *cobra.Command, args []string) {
		d := driver.NewDriver(nodeEndpoint)
		s := server.New(d, nil, d)
		if err := s.Run(nodeEndpoint); err != nil {
			klog.Fatalf("error running server: %v", err)
		}
	},
}

func init() {
	nodeCmd.PersistentFlags().StringVar(&nodeEndpoint, "endpoint", "unix:///tmp/csi.sock", "The endpoint for the CSI driver.")
	rootCmd.AddCommand(nodeCmd)
}
