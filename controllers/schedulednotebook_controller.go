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
	"bytes"
	"context"
	"fmt"

	kbatchv1 "k8s.io/api/batch/v1"
	kcorev1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apiTypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	papermillv1 "github.com/andersonreyes/papermillcontroller/api/v1"
)

// ScheduledNotebookReconciler reconciles a ScheduledNotebook object
type ScheduledNotebookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=papermill.papermill.dev,resources=schedulednotebooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=papermill.papermill.dev,resources=schedulednotebooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=papermill.papermill.dev,resources=schedulednotebooks/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScheduledNotebook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ScheduledNotebookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var notebookWorkflows papermillv1.ScheduledNotebookList
	if err := r.List(ctx, &notebookWorkflows, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to fetch ScheduledNotebookList")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info(fmt.Sprintf("Notebook workflows: %d", len(notebookWorkflows.Items)))

	// Check all notebooks have corresponding CronJob
	for _, notebook := range notebookWorkflows.Items {
		// TODO: We should check if there is an existing cronjob already for this notebook.
		var cronJob kbatchv1.CronJob

		if err := r.Get(ctx, req.NamespacedName, &cronJob); err != nil {
			cronJob, err := createCronJob(notebook)
			if err != nil {
				log.Error(err, fmt.Sprintf("Failed to construct CronJob from notebook %s", notebook.Spec.Name))
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			if err := r.Create(ctx, cronJob); err != nil {
				log.Error(err, fmt.Sprintf("Failed to created cronjob in cluster %s", notebook.Spec.Name))
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			log.Info(fmt.Sprintf("Created ScheduledNotebook %s", notebook.Name))
		}

	}

	var allManagedCronJobs kbatchv1.CronJobList
	if err := r.List(ctx, &allManagedCronJobs, client.MatchingLabels{
		"app": "ScheduledNotebooks",
	}); err != nil {
		log.Error(err, "Failed to retrieve ScheduledNotebook CronJobs")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, cronJob := range allManagedCronJobs.Items {
		var notebook papermillv1.ScheduledNotebook
		name := apiTypes.NamespacedName{
			Namespace: cronJob.Namespace,
			Name:      cronJob.Name,
		}
		if err := r.Get(ctx, name, &notebook); err != nil {
			// CronJob has no ScheduledNotebook object so just delete it
			if err := r.Delete(ctx, &cronJob); err != nil {
				log.Error(err, fmt.Sprintf("Failed to delete CronJob %s in  namespace %s", cronJob.Name, cronJob.Namespace))
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			log.Info(fmt.Sprintf("Deleted ScheduledNotebook and CronJob %s", cronJob.Name))
		}

	}

	return ctrl.Result{}, nil
}

func parametersAsString(p *map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range *p {
		fmt.Fprintf(b, "-p %s %s", key, value)
	}

	return b.String()
}

func createCronJob(n papermillv1.ScheduledNotebook) (*kbatchv1.CronJob, error) {
	labels := make(map[string]string)
	labels["app"] = "ScheduledNotebooks"

	for key, value := range n.ObjectMeta.Labels {
		labels[key] = value
	}

	papermillParameters := parametersAsString(&n.Spec.Parameters)
	papermillCommand := fmt.Sprintf("papermill %s %s %s && cat %s", papermillParameters, n.Spec.InputNotebook, n.Spec.OutputNotebook, n.Spec.OutputNotebook)
	cronJob := &kbatchv1.CronJob{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:      n.ObjectMeta.Name,
			Namespace: n.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: kbatchv1.CronJobSpec{
			Schedule:          n.Spec.Schedule,
			ConcurrencyPolicy: "Forbid",
			JobTemplate: kbatchv1.JobTemplateSpec{
				Spec: kbatchv1.JobSpec{
					Template: kcorev1.PodTemplateSpec{
						Spec: kcorev1.PodSpec{
							RestartPolicy: "Never",
							Containers: []kcorev1.Container{
								{
									Name:            n.Name,
									ImagePullPolicy: "IfNotPresent",
									Image:           n.Spec.DockerImage,
									Args: []string{
										papermillCommand,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return cronJob, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScheduledNotebookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&papermillv1.ScheduledNotebook{}).
		Complete(r)
}
