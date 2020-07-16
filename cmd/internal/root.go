package cmd

import (
	"fmt"
	"os"

	eirinix "github.com/SUSE/eirinix"
	ingress "github.com/mudler/eirini-ingress/extensions/ingress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

var cfgFile string
var kubeconfig string
var namespace string
var rootCmd = &cobra.Command{
	Use:   "eirini-ingress",
	Short: "eirini-ingress creates ingress and services for apps pushed in Cloud Foundry",
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("kubeconfig", cmd.Flags().Lookup("kubeconfig"))
		viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))

		viper.BindEnv("kubeconfig")
		viper.BindEnv("namespace", "NAMESPACE")
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		ns := viper.GetString("namespace")
		filter := false
		opts := eirinix.ManagerOptions{
			Namespace:           ns,
			KubeConfig:          viper.GetString("kubeconfig"),
			OperatorFingerprint: "eirini-ingress", // Not really used for now, but setting it up for future
			FilterEiriniApps:    &filter,
		}
		x := eirinix.NewManager(opts)
		x.GetLogger().Info("Starting watcher in ", x.GetManagerOptions().Namespace)
		x.GetLogger().Info(" Kubeconfig ", x.GetManagerOptions().KubeConfig)

		// Getting start RV for the specific namespace
		client, err := x.GetKubeClient()
		if err != nil {
			x.GetLogger().Error((err.Error()))
		}

		lw := cache.NewListWatchFromClient(client.RESTClient(), "pods", ns, fields.Everything())
		list, err := lw.List(metav1.ListOptions{})
		if err != nil {
			x.GetLogger().Error((err.Error()))
			os.Exit(1)

		}

		metaObj, err := meta.ListAccessor(list)
		if err != nil {
			x.GetLogger().Error((err.Error()))
			os.Exit(1)

		}

		opts.WatcherStartRV = metaObj.GetResourceVersion()
		x.SetManagerOptions(opts)
		x.AddWatcher(ingress.NewPodWatcher())
		err = x.Watch()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "eirini", "Namespace to watch for Eirini apps")
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to a kubeconfig, not required in-cluster")
}
