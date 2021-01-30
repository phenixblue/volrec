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

// PersistentVolumeReconciler reconciles a PersistentVolume object
type PersistentVolumeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// VolumeMap maps a Kubernetes Persistent Volume, the associated Volume Claim, and the
// Namespace that it's assocaited with
type VolumeMap struct {
	pvName           string
	pvClaimKind      string
	pvClaimName      string
	pvClaimNamespace string
	nsOwner          string
}

// buildNamespaceMap Builds mapping of PV -> PVC -> Namespace and associated owner
func buildNamespaceMap(ctx context.Context, r *PersistentVolumeReconciler, log logr.Logger, pv corev1.PersistentVolume, ownerLabel string) string {
	var (
		ns corev1.Namespace
	)

	nsOwner := ""

	if err := r.Get(ctx, client.ObjectKey{Name: pv.Spec.ClaimRef.Namespace}, &ns); err != nil {
		log.Error(err, "unable to fetch namespace")
		return ""
	}

	nsOwner = ns.GetLabels()[config.VolrecConfig.OwnerLabel]
	return nsOwner

}

// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;update;patch

// Reconcile reconciles Kubernetes Persistent Volumes for the Volume Reclaim Controller (VRC) Controller
func (r *PersistentVolumeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pv", req.NamespacedName)

	var (
		pv  corev1.PersistentVolume
		pvc corev1.PersistentVolumeClaim
		// Commenting out until i solve for no shell in controller image
		//reclaimPolicyLabel string = viper.GetString("storage.reclaim.label")
		//ownerLabel         string = viper.GetString("owner.label")
		//ownerSet           bool   = viper.GetBool("owner.set-owner")
		pvMap VolumeMap
	)

	if err := r.Get(ctx, req.NamespacedName, &pv); err != nil {
		log.Error(err, "unable to fetch PV")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Need to figure out how to filter this on delete events
	if err := r.Get(ctx, client.ObjectKey{Name: pv.Spec.ClaimRef.Name, Namespace: pv.Spec.ClaimRef.Namespace}, &pvc); err != nil {
		//log.Error(err, "unable to fetch PVC", "namespace", pvc.Namespace, "pvc", pvc.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	pvcLabels := pvc.GetLabels()
	reclaimPolicyFromPVCLabel := pvcLabels[config.VolrecConfig.ReclaimPolicyLabel]
	log.Info("Reconciling PV", "policy-from-label", reclaimPolicyFromPVCLabel)

	/* Don't think we need this section....only valid for direct edits to PV which should probably be ignored to allow for admin override
	// if value in label does not match value on PV, set it
	if fmt.Sprintf("%v", pv.Spec.PersistentVolumeReclaimPolicy) != reclaimPolicyFromPVCLabel {
		log.Info("Setting reclaim policy to match PVC label", "policy-from-pvc-label", reclaimPolicyFromPVCLabel, "policy-from-pv", pv.Spec.PersistentVolumeReclaimPolicy)
		// Update the reclaim policy from label value
		pv.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimPolicy(reclaimPolicyFromPVCLabel)
	} else {
		log.Info("Reclaim policy on PV already matches PVC label", "policy-from-pvc-label", reclaimPolicyFromPVCLabel, "policy-from-pv", pv.Spec.PersistentVolumeReclaimPolicy)
	}
	*/

	// if owner label is enabled and does not already exist, set it
	if config.VolrecConfig.OwnerSet == true || config.VolrecConfig.NsSet == true {

		pvMap.nsOwner = buildNamespaceMap(ctx, r, log, pv, config.VolrecConfig.OwnerLabel)
		pvMap.pvName = pv.Name
		pvMap.pvClaimKind = pv.Spec.ClaimRef.Kind
		pvMap.pvClaimName = pv.Spec.ClaimRef.Name
		pvMap.pvClaimNamespace = pv.Spec.ClaimRef.Namespace

		// Set Owner Label
		if config.VolrecConfig.OwnerSet == true {
			if pvMap.nsOwner == pv.GetLabels()[config.VolrecConfig.OwnerLabel] {
				log.Info("Owner label already set on PV", "owner-label", config.VolrecConfig.OwnerLabel, "owner", pvMap.nsOwner)
			} else if pvMap.nsOwner == "" {
				log.Info("Owner label isn't set or is blank", "owner-label", config.VolrecConfig.OwnerLabel)
			} else {
				log.Info("Setting Owner label", "owner-label", config.VolrecConfig.OwnerLabel, "owner", pvMap.nsOwner)
				if len(pv.Labels) == 0 {
					pv.Labels = make(map[string]string)
				}
				pv.Labels[config.VolrecConfig.OwnerLabel] = pvMap.nsOwner
			}
		}

		// Set Owning Namespace Label
		if config.VolrecConfig.NsSet == true {
			if pvMap.pvClaimNamespace == pv.GetLabels()[config.VolrecConfig.NsLabel] {
				log.Info("Owning Namespace label already set on PV", "ns-label", config.VolrecConfig.NsLabel, "namespace", pvMap.pvClaimNamespace)
			} else if pvMap.pvClaimNamespace == "" {
				log.Info("Namespace not set in claimRef on PV")
			} else {
				log.Info("Setting owning Namespace label", "ns-label", config.VolrecConfig.NsLabel, "namespace", pvMap.pvClaimNamespace)
				if len(pv.Labels) == 0 {
					pv.Labels = make(map[string]string)
				}
				pv.Labels[config.VolrecConfig.NsLabel] = pvMap.pvClaimNamespace
			}
		}
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
func (r *PersistentVolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolume{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Ignore updates to CR status in which case metadata.Generation does not change
				return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				//return !e.DeleteStateUnknown
				return false
			},
		}).
		Complete(r)
}
