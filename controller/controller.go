// Copyright 2022 Ahmet Alp Balkan

package main

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kptv1alpha1 "kpt.dev/resourcegroup/apis/kpt.dev/v1alpha1"
	"sigs.k8s.io/cli-utils/pkg/common"
	cr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"bundle-deployment-controller/apis/v1alpha1"
	"bundle-deployment-controller/controller/deployer"
	"bundle-deployment-controller/controller/release"
)

const deploymentCleanupFinalizer = `ahmet.dev/deployment-cleanup`

type reconciler struct {
	client client.Client
}

func (r *reconciler) Reconcile(ctx context.Context, req cr.Request) (cr.Result, error) {
	log := log.FromContext(ctx)
	log.Info("start reconcile")
	defer log.Info("end reconcile")

	var obj v1alpha1.BundleDeployment
	err := r.client.Get(ctx, req.NamespacedName, &obj)
	if err != nil {
		return cr.Result{}, client.IgnoreNotFound(err)
	}

	if modified, err := r.reconcileFinalizer(ctx, &obj); err != nil {
		return cr.Result{}, err
	} else if modified {
		return cr.Result{}, nil
	}

	if err := r.reconcileResourceGroup(ctx,
		inventoryObjFor(&obj),
		inventoryIDFor(&obj)); err != nil {
		return cr.Result{}, err
	}

	return cr.Result{}, r.reconcileDeployment(ctx, &obj)
}

func (r *reconciler) reconcileResourceGroup(ctx context.Context, key client.ObjectKey, inventoryID string) error {
	log := log.FromContext(ctx)
	var rg kptv1alpha1.ResourceGroup
	err := r.client.Get(ctx, key, &rg)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}
	log.Info("creating inventory object")

	rg = kptv1alpha1.ResourceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
			Labels: map[string]string{
				common.InventoryLabel: inventoryID},
		},
	}
	return r.client.Create(ctx, &rg)
}

func (r *reconciler) reconcileFinalizer(ctx context.Context,
	obj *v1alpha1.BundleDeployment) (modified bool, err error) {
	log := log.FromContext(ctx)

	if obj.DeletionTimestamp.IsZero() {
		// object is not being deleted
		if controllerutil.AddFinalizer(obj, deploymentCleanupFinalizer) {
			log.Info("object missing finalizer")
			return true, r.client.Update(ctx, obj)
		}
		return false, nil
	}

	if !controllerutil.ContainsFinalizer(obj, deploymentCleanupFinalizer) {
		return false, nil
	}

	out, err := deployer.Destroy(ctx, inventoryObjFor(obj), inventoryIDFor(obj))
	if err != nil {
		return false, fmt.Errorf("failed to destroy deployment: %w", err)
	}
	log.Info("destroy events", "out", out)

	controllerutil.RemoveFinalizer(obj, deploymentCleanupFinalizer)
	return true, r.client.Update(ctx, obj)
}

func (r *reconciler) reconcileDeployment(ctx context.Context, obj *v1alpha1.BundleDeployment) error {
	log := log.FromContext(ctx)
	bundle, ok := release.ReleaseBundles[obj.Spec.ReleaseBundle]
	if !ok {
		return fmt.Errorf("unrecognized release bundle %q", obj.Spec.ReleaseBundle)
	}

	releaseChecksum, err := bundle.Checksum()
	if err != nil {
		return fmt.Errorf("failed to checksum release bundle: %w", err)
	}

	if !needsReapply(obj, releaseChecksum) {
		log.Info("deployment is up to date")
		return nil
	}

	manifests, err := deployer.BuildManifests(bundle,
		deployer.WithManifestNamePrefix(obj.GetName()),
		deployer.WithReplicasPatch(obj.Spec.Replicas))
	if err != nil {
		return fmt.Errorf("failed to compile deployment manifests: %w", err)
	}

	// TODO use context.WithTimeout to set waiting on deployment readiness
	deployOutput, err := deployer.Deploy(ctx, inventoryObjFor(obj), inventoryIDFor(obj),
		manifests)
	log.Info("output", "msg", deployOutput)
	if err != nil {
		return fmt.Errorf("applying manifests and waiting for readiness failed: %w", err)
	}

	// Update status
	obj.Status.ObservedGeneration = obj.GetGeneration()
	obj.Status.LastAppliedReleaseChecksum = releaseChecksum
	obj.Status.Ready = true
	return r.client.Status().Update(ctx, obj)
}

func needsReapply(obj *v1alpha1.BundleDeployment, currentReleaseChecksum string) bool {
	return obj.GetGeneration() != obj.Status.ObservedGeneration ||
		obj.Status.LastAppliedReleaseChecksum != currentReleaseChecksum
}

func inventoryObjFor(rc *v1alpha1.BundleDeployment) client.ObjectKey {
	return client.ObjectKeyFromObject(rc)
}

func inventoryIDFor(rc *v1alpha1.BundleDeployment) string {
	return string(rc.GetObjectMeta().GetUID())
}
