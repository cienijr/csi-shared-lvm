package cmd

import (
	"flag"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var rootCmd = &cobra.Command{
	Use:   "csi-shared-lvm",
	Short: "A Kubernetes CSI Driver for shared storage based on LVM",
	Long:  `A CSI driver for Kubernetes that enables provisioning and management of LVM volumes over a shared block storage device.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		klog.Fatalf("error executing root command: %v", err)
	}
}

func init() {
	var fs flag.FlagSet
	klog.InitFlags(&fs)
	rootCmd.PersistentFlags().AddGoFlagSet(&fs)
}
