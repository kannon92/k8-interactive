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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	interactive "k8.interactive.job/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		name := fmt.Sprintf("%s", interactiveJob.Name)

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Name:        name,
				Namespace:   interactiveJob.Namespace,
			},
			Spec: *interactiveJob.Spec.JobTemplate.Spec.DeepCopy(),
		}

		for k, v := range interactiveJob.Spec.JobTemplate.Labels {
			job.Labels[k] = v
		}

		for k, v := range interactiveJob.Spec.JobTemplate.Annotations {
			job.Annotations[k] = v
		}
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
		return ctrl.Result{}, nil
	}

	// ...and create it on the cluster
	if err := r.Create(ctx, job); err != nil {
		log.Error(err, "unable to create Job for InteractiveJob", "job", job)
		return ctrl.Result{}, err
	}

	constructServiceForInteractiveJob := func(interactiveJob *interactive.InteractiveJob) (*corev1.Service, error) {
		name := fmt.Sprintf("%s-service", job.Name)

		// Need to get port and name from job

		servicePorts := make([]corev1.ServicePort, 1)
		for _, v := range interactiveJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
			for _, port := range v.Ports {
				servicePorts = append(servicePorts, corev1.ServicePort{Port: port.ContainerPort})
			}
		}
		if len(servicePorts) == 0 {
			return nil, fmt.Errorf("The Job Template needs to have container ports")
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Name:        name,
				Namespace:   interactiveJob.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Type:  "NodePort",
				Ports: servicePorts,
			},
		}
		for k, v := range interactiveJob.Spec.JobTemplate.Annotations {
			service.Annotations[k] = v
		}

		for k, v := range interactiveJob.Spec.JobTemplate.Labels {
			service.Labels[k] = v
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
	}
	if serviceErr := r.Create(ctx, service); serviceErr != nil {
		log.Error(serviceErr, "unable to create Service for InteractiveJob", "service", service)
	}

	log.V(1).Info("created Job for InteractiveJob run", "job", interactiveJob)
	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InteractiveJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interactive.InteractiveJob{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
