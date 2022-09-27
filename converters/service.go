package converters

import (
	"fmt"

	interactive "k8.interactive.job/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateServiceFromInteractiveJob(interactiveJob *interactive.InteractiveJob) (*corev1.Service, error) {
	serviceName := fmt.Sprintf("%s-service", interactiveJob.Name)
	servicePorts := make([]corev1.ServicePort, 1)
	for index, port := range interactiveJob.Spec.Service.Ports {
		servicePorts[index] = corev1.ServicePort{Port: port}
	}
	if len(servicePorts) == 0 {
		return nil, fmt.Errorf("The Job Template needs to have container ports")
	}
	selector := make(map[string]string)
	selector["app"] = interactiveJob.Name
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      interactiveJob.Labels,
			Annotations: interactiveJob.Annotations,
			Name:        serviceName,
			Namespace:   interactiveJob.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     "NodePort",
			Selector: selector,
			Ports:    servicePorts,
		},
	}
	return service, nil
}

func GenerateJobFromInteractiveJob(interactiveJob *interactive.InteractiveJob) *batchv1.Job {

	var backoffLimit int32 = 1
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      interactiveJob.Labels,
			Annotations: interactiveJob.Annotations,
			Name:        interactiveJob.Name,
			Namespace:   interactiveJob.Namespace,
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: interactiveJob.Spec.JobTemplate.ActiveDeadlineSeconds,
			BackoffLimit:          &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: interactiveJob.Labels,
				},
				Spec: *interactiveJob.Spec.JobTemplate.Template.Spec.DeepCopy(),
			},
		},
	}

	return job
}
