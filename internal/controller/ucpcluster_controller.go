package controller

import (
    "context"
    "fmt"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    ucpv1alpha1 "github.com/raphaeljuwe/ucp-operator/api/v1alpha1"
)

// UCPClusterReconciler reconciles a UCPCluster object
type UCPClusterReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ucp.example.com,resources=ucpclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ucp.example.com,resources=ucpclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ucp.example.com,resources=ucpclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *UCPClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Fetch the UCPCluster instance
    ucpCluster := &ucpv1alpha1.UCPCluster{}
    err := r.Get(ctx, req.NamespacedName, ucpCluster)
    if err != nil {
        if errors.IsNotFound(err) {
            // Request object not found, could have been deleted after reconcile request.
            // Return and don't requeue
            log.Info("UCPCluster resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        // Error reading the object - requeue the request.
        log.Error(err, "Failed to get UCPCluster")
        return ctrl.Result{}, err
    }

    // Check if the deployment already exists, if not create a new one
    found := &appsv1.Deployment{}
    err = r.Get(ctx, types.NamespacedName{Name: ucpCluster.Name, Namespace: ucpCluster.Namespace}, found)
    if err != nil && errors.IsNotFound(err) {
        // Define a new deployment
        dep := r.deploymentForUCPCluster(ucpCluster)
        log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
        err = r.Create(ctx, dep)
        if err != nil {
            log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
            return ctrl.Result{}, err
        }
        // Deployment created successfully - return and requeue
        return ctrl.Result{Requeue: true}, nil
    } else if err != nil {
        log.Error(err, "Failed to get Deployment")
        return ctrl.Result{}, err
    }

    // Ensure the deployment size is the same as the spec
    size := ucpCluster.Spec.Size
    if *found.Spec.Replicas != size {
        found.Spec.Replicas = &size
        err = r.Update(ctx, found)
        if err != nil {
            log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
            return ctrl.Result{}, err
        }
        // Spec updated - return and requeue
        return ctrl.Result{Requeue: true}, nil
    }

    // Update the UCPCluster status with the pod names
    // List the pods for this UCPCluster's deployment
    podList := &corev1.PodList{}
    listOpts := []client.ListOption{
        client.InNamespace(ucpCluster.Namespace),
        client.MatchingLabels(labelsForUCPCluster(ucpCluster.Name)),
    }
    if err = r.List(ctx, podList, listOpts...); err != nil {
        log.Error(err, "Failed to list pods", "UCPCluster.Namespace", ucpCluster.Namespace, "UCPCluster.Name", ucpCluster.Name)
        return ctrl.Result{}, err
    }
    podNames := getPodNames(podList.Items)

    // Update status.Nodes if needed
    if !reflect.DeepEqual(podNames, ucpCluster.Status.Nodes) {
        ucpCluster.Status.Nodes = podNames
        err := r.Status().Update(ctx, ucpCluster)
        if err != nil {
            log.Error(err, "Failed to update UCPCluster status")
            return ctrl.Result{}, err
        }
    }

    return ctrl.Result{}, nil
}

// deploymentForUCPCluster returns a UCPCluster Deployment object
func (r *UCPClusterReconciler) deploymentForUCPCluster(m *ucpv1alpha1.UCPCluster) *appsv1.Deployment {
    ls := labelsForUCPCluster(m.Name)
    replicas := m.Spec.Size

    dep := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      m.Name,
            Namespace: m.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: ls,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: ls,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{{
                        Image: "ucpimage:v1", // Replace with actual UCP image
                        Name:  "ucp",
                        Ports: []corev1.ContainerPort{{
                            ContainerPort: 443,
                            Name:          "ucp",
                        }},
                    }},
                },
            },
        },
    }
    // Set UCPCluster instance as the owner and controller
    ctrl.SetControllerReference(m, dep, r.Scheme)
    return dep
}

// labelsForUCPCluster returns the labels for selecting the resources
// belonging to the given UCPCluster CR name.
func labelsForUCPCluster(name string) map[string]string {
    return map[string]string{"app": "ucpcluster", "ucpcluster_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
    var podNames []string
    for _, pod := range pods {
        podNames = append(podNames, pod.Name)
    }
    return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *UCPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&ucpv1alpha1.UCPCluster{}).
        Owns(&appsv1.Deployment{}).
        Complete(r)
}
