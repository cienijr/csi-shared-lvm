package cmd

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/component-base/config"
	"k8s.io/component-base/config/options"
	"k8s.io/component-base/config/validation"
	"k8s.io/klog/v2"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cienijr/csi-shared-lvm/pkg/driver"
	"github.com/cienijr/csi-shared-lvm/pkg/server"
)

var (
	controllerEndpoint   string
	leaderElectionConfig = config.LeaderElectionConfiguration{
		LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
		RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
		RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
		ResourceLock:      resourcelock.LeasesResourceLock,
		ResourceName:      "csi-shared-lvm-controller",
		ResourceNamespace: "kube-system",
	}
)

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Runs the CSI controller plugin",
	Long:  `Runs the CSI controller plugin.`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		err := validation.ValidateLeaderElectionConfiguration(&leaderElectionConfig, field.NewPath("leaderElect")).ToAggregate()

		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		if !leaderElectionConfig.LeaderElect {
			klog.Info("leader election is disabled, starting gRPC server directly")
			runServer()
			return
		}

		klog.Info("leader election is enabled")
		hostname, err := os.Hostname()
		if err != nil {
			klog.Fatalf("failed to get hostname: %v", err)
		}

		cfg := ctrl.GetConfigOrDie()
		client := clientset.NewForConfigOrDie(cfg)
		lock := &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      leaderElectionConfig.ResourceName,
				Namespace: leaderElectionConfig.ResourceNamespace,
			},
			Client: client.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: hostname,
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:          lock,
			LeaseDuration: leaderElectionConfig.LeaseDuration.Duration,
			RenewDeadline: leaderElectionConfig.RenewDeadline.Duration,
			RetryPeriod:   leaderElectionConfig.RetryPeriod.Duration,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					klog.Info("became leader, starting gRPC server")
					runServer()
				},
				OnStoppedLeading: func() {
					klog.Info("stopped leading")
					cancel()
					klog.Fatalf("leader election lost")
				},
				OnNewLeader: func(identity string) {
					if identity == hostname {
						return
					}
					klog.Infof("new leader elected: %s", identity)
				},
			},
			ReleaseOnCancel: true,
		})
	},
}

func runServer() {
	d := driver.NewDriver(controllerEndpoint)
	s := server.New(d, d, nil)
	if err := s.Run(controllerEndpoint); err != nil {
		klog.Fatalf("error running server: %v", err)
	}
}

func init() {
	var fs flag.FlagSet
	controllerCmd.PersistentFlags().StringVar(&controllerEndpoint, "endpoint", "unix:///tmp/csi.sock", "The endpoint for the CSI driver.")
	options.BindLeaderElectionFlags(&leaderElectionConfig, controllerCmd.PersistentFlags())
	ctrl.RegisterFlags(&fs)
	controllerCmd.PersistentFlags().AddGoFlagSet(&fs)
	rootCmd.AddCommand(controllerCmd)
}
