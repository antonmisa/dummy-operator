/*
Copyright 2023.

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
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dummyv1alpha1 "github.com/antonmisa/dummy-operator/api/v1alpha1"
)

const (
	containerImageName string = "nginx"
	containerName      string = "nginx"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Define logger with default parameters name and namespace
	logger := log.Log.WithValues("Name", req.Name, "Namespace", req.Namespace)

	// Create new dummy instance
	dummy := &dummyv1alpha1.Dummy{}

	// Retrive dummy instance information
	err := r.Get(ctx, req.NamespacedName, dummy)
	if err != nil {
		// Check if not found then it's a deletion request,
		// everything is OK, because all child objects was deleted
		if errors.IsNotFound(err) {
			logger.Info("Dummy was deleted")

			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: dummy.Name, Namespace: dummy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Deployment not found, so let's create a new one
		// Define a new deployment
		dep := r.deploymentForDummy(dummy)

		logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

		// Create a new deployment
		err = r.Create(ctx, dep)
		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}

		// Deployment created successfully - return and requeue
		//return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// List the pods for this dummy's deployment
	phasesList, err := r.getPodsPhases(ctx, dummy)
	if err != nil {
		logger.Error(err, "Failed to retrieve Dummy pods phases")
		return ctrl.Result{}, err
	}

	// For simplicity convert phasesList to string
	phases := convertToString(phasesList)

	// Update status if needed
	//	if (dummy.Status.PodStatus != phases) || (dummy.Status.SpecEcho != dummy.Spec.Message) {
	logger.Info(fmt.Sprintf("message: %s", dummy.Spec.Message))

	dummy.Status.SpecEcho = dummy.Spec.Message
	dummy.Status.PodStatus = phases

	// Updating status of dummy
	err = r.updateDummyStatus(ctx, dummy)
	if err != nil {
		logger.Error(err, "Failed to update Dummy status")
		return ctrl.Result{}, err
	}
	//	}
	return ctrl.Result{}, nil
}

// getPodsStatus retrieve status of dummies pods
func (r *DummyReconciler) getPodsPhases(ctx context.Context, d *dummyv1alpha1.Dummy) ([]string, error) {
	// Create list of pods
	podList := &corev1.PodList{}

	// Define where and what pods we will find
	listOpts := []client.ListOption{
		client.InNamespace(d.Namespace),
		client.MatchingLabels(labelsForDummy(d.Name)),
	}

	// Getting list of pods with defined criteria
	if err := r.List(ctx, podList, listOpts...); err != nil {
		return []string{}, err
	}

	// Output slice of pods phases
	podStatuses := make([]string, 0, len(podList.Items))

	// Fill the slice
	for _, pod := range podList.Items {
		podStatuses = append(podStatuses, string(pod.Status.Phase))
	}

	return podStatuses[:len(podStatuses):len(podStatuses)], nil
}

// Update status of Dummy instance
func (r *DummyReconciler) updateDummyStatus(ctx context.Context, d *dummyv1alpha1.Dummy) error {
	err := r.Status().Update(ctx, d)
	if err != nil {
		return err
	}

	return nil
}

// Deployment creation for dummy instance with nginx image container
func (r *DummyReconciler) deploymentForDummy(d *dummyv1alpha1.Dummy) *appsv1.Deployment {
	// Define labels for dummy pod
	ls := labelsForDummy(d.Name)

	// Create deployment with nginx image container
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: containerImageName,
						Name:  containerName,
					}},
				},
			},
		},
	}

	// Set Dummy instance as the owner and controller
	ctrl.SetControllerReference(d, dep, r.Scheme)

	return dep
}

func convertToString(phases []string) string {
	sort.Strings(phases)

	phases = unique(phases)

	return strings.Join(phases, ", ")
}

func unique[T comparable](s []T) []T {
	if len(s) < 2 {
		return s
	}

	j := 0
	for i := 1; i < len(s); i++ {
		if s[j] != s[i] {
			j++

			if j < i {
				s[j] = s[i]

				for k := i + 1; k < len(s); k++ {
					if s[j] != s[k] {
						j++
						s[j] = s[k]
					}
				}
				break
			}
		}
	}
	return s[: j+1 : j+1]
}

// labelsForDummy returns the labels for selecting the resources
// belonging to the given dummies CR name.
func labelsForDummy(name string) map[string]string {
	return map[string]string{"app": "dummy", "dummy_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dummyv1alpha1.Dummy{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
