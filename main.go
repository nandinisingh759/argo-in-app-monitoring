/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	argoinappiov1 "monitoring/api/v1"
	"monitoring/controllers"

	"github.com/argoproj/notifications-engine/pkg/api"
	"github.com/argoproj/notifications-engine/pkg/controller"
	"github.com/argoproj/notifications-engine/pkg/services"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	//+kubebuilder:scaffold:imports
)

var (
	scheme       = runtime.NewScheme()
	setupLog     = ctrl.Log.WithName("setup")
	clientConfig clientcmd.ClientConfig
)

/** Notifications Controller **/
func addK8SFlagsToCmd() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}

func notificationsController() {
	// Get Kubernetes REST Config and current Namespace so we can talk to Kubernetes
	clientConfig = addK8SFlagsToCmd()
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		log.Fatalf("Failed to get Kubernetes config")
	}
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		log.Fatalf("Failed to get namespace from Kubernetes config")
	}

	// Create ConfigMap and Secret informer to access notifications configuration
	informersFactory := informers.NewSharedInformerFactoryWithOptions(
		kubernetes.NewForConfigOrDie(restConfig),
		time.Minute,
		informers.WithNamespace(namespace))
	secrets := informersFactory.Core().V1().Secrets().Informer()
	configMaps := informersFactory.Core().V1().ConfigMaps().Informer()

	// Create "Notifications" API factory that handles notifications processing
	notificationsFactory := api.NewFactory(api.Settings{
		ConfigMapName: "test-notifs",
		SecretName:    "analysis-notifications-secret",
		InitGetVars: func(cfg *api.Config, configMap *v1.ConfigMap, secret *v1.Secret) (api.GetVars, error) {
			return func(obj map[string]interface{}, dest services.Destination) map[string]interface{} {
				return map[string]interface{}{"run": obj}
			}, nil
		},
	}, namespace, secrets, configMaps)

	// Create notifications controller that handles Kubernetes resources processing
	runClient := dynamic.NewForConfigOrDie(restConfig).Resource(schema.GroupVersionResource{
		Group: "argo-in-app.io", Version: "v1", Resource: "metricruns",
	})

	runInformer := cache.NewSharedIndexInformer(&cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			runMetrics, err := runClient.List(context.Background(), options)
			return runMetrics, err
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			runWatch, err := runClient.Watch(context.Background(), metav1.ListOptions{})
			return runWatch, err
		},
	}, &unstructured.Unstructured{}, time.Minute, cache.Indexers{})

	ctrl1 := controller.NewController(runClient, runInformer, notificationsFactory)
	// Start informers and controller
	go informersFactory.Start(context.Background().Done())
	go runInformer.Run(context.Background().Done())
	if !cache.WaitForCacheSync(context.Background().Done(), secrets.HasSynced, configMaps.HasSynced, runInformer.HasSynced) {
		log.Fatalf("Failed to synchronize informers")
	}

	go ctrl1.Run(10, context.Background().Done())

}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(argoinappiov1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "07530747.argo-in-app.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.InAppMetricReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "InAppMetric")
		os.Exit(1)
	}
	if err = (&controllers.MetricRunReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MetricRun")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	notificationsController()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
