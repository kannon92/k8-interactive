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

package controllers

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	interactive "k8.interactive.job/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	converters "k8.interactive.job/converters"
)

// InteractiveJobReconciler reconciles a InteractiveJob object
type InteractiveJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.k8.interactive.job,resources=interactivejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.k8.interactive.job,resources=interactivejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.k8.interactive.job,resources=interactivejobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InteractiveJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *InteractiveJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var interactiveJob interactive.InteractiveJob
	if err := r.Get(ctx, req.NamespacedName, &interactiveJob); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	constructJobForInteractiveJob := func(interactiveJob *interactive.InteractiveJob) (*batchv1.Job, error) {
		job := converters.GenerateJobFromInteractiveJob(interactiveJob)

		if err := ctrl.SetControllerReference(interactiveJob, job, r.Scheme); err != nil {
			return nil, err
		}

		return job, nil
	}
	// actually make the job...
	job, err := constructJobForInteractiveJob(&interactiveJob)
	if err != nil {
		log.Error(err, "unable to construct job from template")
		// don't bother requeuing until we get a change to the spec
		return ctrl.Result{}, err
	}
	// ...and create it on the cluster
	if err := r.Create(ctx, job); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Error(err, "unable to create Job for InteractiveJob", "job", job)
			return ctrl.Result{}, err
		}
		emptyJob := batchv1.Job{}
		if err := r.Get(ctx, req.NamespacedName, &emptyJob); err != nil {
			log.Error(err, "Cant find job corresponding to InteractiveJob")
			return ctrl.Result{}, err
		}
		status := "FAILED"
		if emptyJob.Status.Active > 0 {
			status = "RUNNING"
		} else if emptyJob.Status.Succeeded > 0 {
			status = "SUCCEEDED"
		}
		interactiveJob.Status.JobStatus = status
		if err := r.Update(ctx, &interactiveJob); err != nil {
			log.Error(err, "Unable to update interactive job")
			return ctrl.Result{}, err
		}
	}

	constructServiceForInteractiveJob := func(interactiveJob *interactive.InteractiveJob) (*corev1.Service, error) {
		service, err := converters.GenerateServiceFromInteractiveJob(interactiveJob)
		if err != nil {
			log.Error(err, "Unable to Construct Service")
			return nil, err
		}
		if err := ctrl.SetControllerReference(interactiveJob, service, r.Scheme); err != nil {
			return nil, err
		}

		return service, nil

	}

	// Construct a service object from
	service, serviceErr := constructServiceForInteractiveJob(&interactiveJob)
	if serviceErr != nil {
		log.Error(serviceErr, "Unable to create service for interactive job")
		return ctrl.Result{}, serviceErr
	}
	if serviceErr := r.Create(ctx, service); serviceErr != nil {
		if !apierrors.IsAlreadyExists(serviceErr) {
			log.Error(serviceErr, "unable to create Service for InteractiveJob", "service", service)
			return ctrl.Result{}, serviceErr
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InteractiveJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interactive.InteractiveJob{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
