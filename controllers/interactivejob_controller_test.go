package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	interactive "k8.interactive.job/api/v1"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("InteractiveJob controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		InteractiveJobName      = "test-intjob"
		InteractiveJobNamespace = "default"
		timeout                 = time.Second * 10
		duration                = time.Second * 10
		interval                = time.Millisecond * 250
	)

	Context("When creating Interactive Job", func() {
		It("Should create a job and service object", func() {
			By("By creating a new InteractiveJob")
			ctx := context.Background()
			InteractiveJob := &interactive.InteractiveJob{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "batch.k8.interactive.job/v1",
					Kind:       "InteractiveJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      InteractiveJobName,
					Namespace: InteractiveJobNamespace,
				},
				Spec: interactive.InteractiveJobSpec{
					JobTemplate: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							// For simplicity, we only fill out the required fields.
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									// For simplicity, we only fill out the required fields.
									Containers: []v1.Container{
										{
											Name:  "minimal-notebook",
											Image: "jupyter/tensorflow-notebook:latest",
										},
									},
									RestartPolicy: v1.RestartPolicyOnFailure,
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, InteractiveJob)).Should(Succeed())
			interactiveJobLookup := types.NamespacedName{Name: InteractiveJobName, Namespace: InteractiveJobNamespace}

			intJob := interactive.InteractiveJob{}
			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, interactiveJobLookup, &intJob)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// This should create Job and Service object
			job := batchv1.Job{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, interactiveJobLookup, &job)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			service := v1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, interactiveJobLookup, &service)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

		})
	})
})
