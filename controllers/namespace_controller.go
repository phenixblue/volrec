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

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

// Reconcile reconciles Kubernetes Namespaces for the Volume Reclaim Controller (VRC) Controller
func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ns", req.NamespacedName)

	var (
		ns  corev1.Namespace
		pvs corev1.PersistentVolumeList
	)

	if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ownerFromNSLabel := ns.GetLabels()[config.VolrecConfig.OwnerLabel]
	log.Info("Reconciling NS", "owner", ownerFromNSLabel)

	// if value in label does not match value on PV, set it
	if ownerFromNSLabel == "" {
		log.Info("NS does not have owner label", "namespace", ns.Name, "owner-label", config.VolrecConfig.OwnerLabel)
		return ctrl.Result{}, nil
	}

	if err := r.List(ctx, &pvs, client.MatchingLabels{config.VolrecConfig.NsLabel: ns.Name}); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("No PV's associated with NS", "namespace", ns.Name, "owner-label", config.VolrecConfig.OwnerLabel)
			return ctrl.Result{}, nil
		}
	}

	for _, pv := range pvs.Items {
		if pv.Labels[config.VolrecConfig.OwnerLabel] == ns.Labels[config.VolrecConfig.OwnerLabel] {
			log.Info("NS Owner on PV already matches NS label", "owner-label", config.VolrecConfig.OwnerLabel, "ns-label-value", ownerFromNSLabel, "pv", pv.Name)
			continue
		}

		log.Info("Setting NS Owner label on PV", "owner-label", config.VolrecConfig.OwnerLabel, "ns-label-value", ownerFromNSLabel, "pv", pv.Name, "pv-label-value", pv.Labels[config.VolrecConfig.OwnerLabel])

		pv.Labels[config.VolrecConfig.OwnerLabel] = ownerFromNSLabel

		// Update Persistent Volume
		err := r.Update(context.TODO(), &pv)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("could not update PV: %+v", err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager adds a Kubernetes controller instance to a Controller Manager
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Ignore updates to CR status in which case metadata.Generation does not change
				return e.MetaOld.GetLabels()[config.VolrecConfig.OwnerLabel] != e.MetaNew.GetLabels()[config.VolrecConfig.OwnerLabel]
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				//return !e.DeleteStateUnknown
				return false
			},
		}).
		Complete(r)
}
