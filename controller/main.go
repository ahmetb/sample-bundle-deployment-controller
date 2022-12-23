// Copyright 2022 Ahmet Alp Balkan

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	kptv1alpha1 "kpt.dev/resourcegroup/apis/kpt.dev/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"bundle-deployment-controller/apis/v1alpha1"
)

func main() {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(kptv1alpha1.AddToScheme(scheme))

	setupLog := controllerruntime.Log.WithName("setup")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	controllerruntime.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := controllerruntime.NewManager(controllerruntime.GetConfigOrDie(), controllerruntime.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
	})
	if err != nil {
		setupLog.Error(err, "error initializing controller manager")
		os.Exit(1)
	}

	r := &reconciler{
		client: mgr.GetClient(),
	}
	if err := builder.ControllerManagedBy(mgr).For(&v1alpha1.BundleDeployment{}).Complete(r); err != nil {
		setupLog.Error(err, "failed to build controller")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(controllerruntime.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "manager failed")
		os.Exit(1)
	}
}
