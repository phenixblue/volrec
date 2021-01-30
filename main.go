/*
Copyright 2021 The WebRoot.

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
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1 "k8s.io/api/core/v1"
	"twr.dev/volrec/controllers"

	c "twr.dev/volrec/pkg/config"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8081", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.String("reclaim-label", "storage.k8s.twr.dev/reclaim-policy", "The label to use for tracking Persistent Volume reclaim policy")
	flag.Bool("set-owner", false, "Toggle whether or not owner information from a given namespace is transfered to the Persistent Volume")
	flag.String("owner-label", "k8s.twr.dev/owner", "The Label to use to set owner information on a Persistent Volume")
	flag.Bool("set-ns", false, "Toggle whether or not to add a label mapping Persistent Volumes back to a namespace")
	flag.String("ns-label", "k8s.twr.dev/owning-namespace", "The label to use for identifying an owning namespace on a Persisent Volume")

	flag.Parse()

	c.InitConfig(setupLog)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "86bf18f9.storage.k8s.twr.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.PersistentVolumeReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("PersistentVolume"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PersistentVolume")
		os.Exit(1)
	}
	if err = (&controllers.PersistentVolumeClaimReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("PersistentVolumeClaim"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PersistentVolumeClaim")
		os.Exit(1)
	}
	if err = (&controllers.NamespaceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Namespace"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if _, err := mgr.GetCache().GetInformer(&corev1.Namespace{}); err != nil {
		setupLog.Error(err, "unable to setup cache", "cache", "Namespace")
		os.Exit(1)
	}
	if _, err := mgr.GetCache().GetInformer(&corev1.PersistentVolume{}); err != nil {
		setupLog.Error(err, "unable to setup cache", "cache", "PersistentVolume")
		os.Exit(1)
	}
	if _, err := mgr.GetCache().GetInformer(&corev1.PersistentVolumeClaim{}); err != nil {
		setupLog.Error(err, "unable to setup cache", "cache", "PersistentVolumeClaim")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
