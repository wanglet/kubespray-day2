/*
Copyright 2021 wanglet.

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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"

	clusterv1alpha1 "github.com/wanglet/kubespray-day2/api/v1alpha1"
)

// KubesprayJobReconciler reconciles a KubesprayJob object
type KubesprayJobReconciler struct {
	KubesprayImage string
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=cluster.wanglet.com,resources=kubesprayjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.wanglet.com,resources=kubesprayjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cluster.wanglet.com,resources=kubesprayjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KubesprayJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *KubesprayJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	job := &clusterv1alpha1.KubesprayJob{}
	if err := r.Get(ctx, req.NamespacedName, job); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("KubesprayJob is deleted.", "req", req)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.createInventory(ctx, req, job); err != nil {
		return ctrl.Result{}, err
	}

	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if err == client.IgnoreNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("Can not get pod with err: %v", err)
		}
		pod = r.generatePodObject(job)
		if err := controllerutil.SetControllerReference(job, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.Create(ctx, pod); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create pod: %s", err)
		}
		msg := fmt.Sprintf("Created pod %v", pod.Name)
		r.Log.Info(msg)
		r.Recorder.Event(job, corev1.EventTypeNormal, "Created", msg)
		return ctrl.Result{}, nil
	}

	if job.Status.Phase != pod.Status.Phase {
		job.Status.Phase = pod.Status.Phase
		r.Client.Status().Update(ctx, job)
		r.Log.Info("update phase.", "req", req, "phase:", job.Status.Phase)
	} else {
		r.Log.Info("Phase is synchronized.", "req", req, "phase:", job.Status.Phase)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubesprayJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha1.KubesprayJob{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
