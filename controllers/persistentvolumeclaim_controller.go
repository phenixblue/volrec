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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"twr.dev/volrec/pkg/config"

	corev1 "k8s.io/api/core/v1"
)

// PersistentVolumeClaimReconciler reconciles a PersistentVolumeClaim object
type PersistentVolumeClaimReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch

// Reconcile reconciles Kubernetes Persistent Volumes Claims for the Volume Reclaim Controller (VRC) Controller
func (r *PersistentVolumeClaimReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pvc", req.NamespacedName)

	var (
		pv  corev1.PersistentVolume
		pvc corev1.PersistentVolumeClaim
	)

	if err := r.Get(ctx, req.NamespacedName, &pvc); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pvc.Spec.VolumeName != "" {

		log.Info("PVC is bound to volume", "volume-name", pvc.Spec.VolumeName)

		if err := r.Get(ctx, client.ObjectKey{Name: pvc.Spec.VolumeName}, &pv); err != nil {
			if apierrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		reclaimPolicyFromPVCLabel := pvc.GetLabels()[config.VolrecConfig.ReclaimPolicyLabel]
		log.Info("Reconciling PV", "policy-from-label", reclaimPolicyFromPVCLabel)

		// if value in label does not match value on PV, set it
		if reclaimPolicyFromPVCLabel == "" {
			log.Info("PVC does not have reclaim policy label", "namespace", pvc.Namespace)
			return ctrl.Result{}, nil
		} else if fmt.Sprintf("%v", pv.Spec.PersistentVolumeReclaimPolicy) != reclaimPolicyFromPVCLabel {
			log.Info("Setting reclaim policy to match PVC label", "pv", pv.Name, "policy-from-pvc-label", reclaimPolicyFromPVCLabel, "policy-from-pv", pv.Spec.PersistentVolumeReclaimPolicy)
			// Update the reclaim policy from label value
			pv.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimPolicy(reclaimPolicyFromPVCLabel)
		} else {
			log.Info("Reclaim policy on PV already matches PVC label", "pv", pv.Name, "policy-from-pvc-label", reclaimPolicyFromPVCLabel, "policy-from-pv", pv.Spec.PersistentVolumeReclaimPolicy)
		}

	} else {
		// Requeue to process PVC again once it is bound to a PV
		log.Info("PVC not bound to volume yet", "namespace", pvc.Namespace)
		return reconcile.Result{Requeue: true}, nil
	}

	// Update Persistent Volume
	err := r.Update(context.TODO(), &pv)
	if err != nil {

		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, fmt.Errorf("could not update PV: %+v", err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager adds a Kubernetes controller instance to a Controller Manager
func (r *PersistentVolumeClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Ignore updates to CR status in which case metadata.Generation does not change
				return e.MetaOld.GetLabels()[config.VolrecConfig.ReclaimPolicyLabel] != e.MetaNew.GetLabels()[config.VolrecConfig.ReclaimPolicyLabel]
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				//return !e.DeleteStateUnknown
				return false
			},
		}).
		Complete(r)
}
