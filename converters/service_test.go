package converters

import (
	"testing"

	"github.com/stretchr/testify/assert"

	interactive "k8.interactive.job/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceConversion(t *testing.T) {
	service, err := GenerateServiceFromInteractiveJob(interactiveJobSample())
	assert.Equal(t, err, nil)
	selector := make(map[string]string)
	selector["app"] = "test"

	expectedService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test",
			Labels:    selector,
		},
		Spec: corev1.ServiceSpec{
			Type:     "NodePort",
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Port: 8888,
				},
			},
		},
	}
	assert.Equal(t, service, expectedService)
}
func TestJobConversion(t *testing.T) {
	selector := make(map[string]string)
	selector["app"] = "test"

	job := GenerateJobFromInteractiveJob(interactiveJobSample())
	expectedJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels:    selector,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:      "test",
							Image:     "busybox",
							Command:   []string{},
							Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{}},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8888,
								},
							},
						},
					},
					NodeSelector: map[string]string{},
				},
			},
		},
	}

	assert.Equal(t, job, expectedJob)
}

func interactiveJobSample() *interactive.InteractiveJob {
	labels := make(map[string]string)
	labels["app"] = "test"
	testJob := &interactive.InteractiveJob{
		ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "test", Labels: labels},
		Spec: interactive.InteractiveJobSpec{
			JobTemplate: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						RestartPolicy: "Never",
						Containers: []corev1.Container{
							{
								Name:      "test",
								Image:     "busybox",
								Command:   []string{},
								Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{}},
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 8888,
									},
								},
							},
						},
						NodeSelector: map[string]string{},
					},
				},
			},
		},
	}
	return testJob
}
