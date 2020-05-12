/*


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
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	webappv1 "moperator/api/v1"
)

// WebsiteReconciler reconciles a Website object
type WebsiteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.nikhil.kubebuilder.io,resources=websites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.nikhil.kubebuilder.io,resources=websites/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;update;patch;create;delete;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

func (r *WebsiteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("website", req.NamespacedName)

	var website webappv1.Website
	if err := r.Get(ctx, req.NamespacedName, &website); err != nil {
		log.Error(err, "unable to fetch webiste")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.

		//log.V(1).Info("website name", "website", website)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//log.V(1).Info("website name", "website", website)
	var childDeployments appsv1.DeploymentList
	if err := r.List(ctx, &childDeployments, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list child Deployments")
		return ctrl.Result{}, err
	}
	var deploymentcreated bool = false
	for _, deploy := range childDeployments.Items {
		//log.V(1).Info("deployment name", "deploy", deploy.Name)
		// add logic to delete old deployments
		if deploy.Name == website.Spec.Text {
			deploymentcreated = true
			break
		}

	}

	if !deploymentcreated {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      website.Spec.Text,
				Namespace: website.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "demo",
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "demo",
						},
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Name:  "web",
								Image: "nginx:1.12",
								Ports: []apiv1.ContainerPort{
									{
										Name:          "http",
										Protocol:      apiv1.ProtocolTCP,
										ContainerPort: 80,
									},
								},
							},
						},
					},
				},
			},
		}
		// ...and create it on the cluster
		//if err := ctrl.SetControllerReference(website, deployment, r.Scheme); err != nil {
		//	log.Error(err, "unable to set owner refernce for webiste")
		//	return ctrl.Result{}, err
		//}
		if err := r.Create(ctx, deployment); err != nil {
			log.Error(err, "unable to create deployment for Website", "website", deployment)
			return ctrl.Result{}, err
		}

		log.V(1).Info("created deployment for website", "deployment", deployment)
	}
	return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, nil
}

//var (
//	jobOwnerKey = ".metadata.controller"
//	apiGVStr    = webappv1.GroupVersion.String()
//)

func (r *WebsiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, jobOwnerKey, func(rawObj runtime.Object) []string {
	// grab the job object, extract the owner...
	//	deploy := rawObj.(*appsv1.Deployment)
	//	owner := metav1.GetControllerOf(deploy)
	//	if owner == nil {
	//		return nil
	//	}
	// ...make sure it's a CronJob...
	//	if owner.APIVersion != apiGVStr || owner.Kind != "Website" {
	//		return nil
	//	}

	// ...and if so, return it
	//	return []string{owner.Name}
	//}); err != nil {
	//	return err
	//}
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Website{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
