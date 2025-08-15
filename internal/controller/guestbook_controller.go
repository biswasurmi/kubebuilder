package controller

import (
    "context"
    "fmt"

    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
	"k8s.io/client-go/util/retry"

    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    webappv1 "my.domain/guestbook/api/v1"
)
// GuestbookReconciler reconciles a Guestbook object
type GuestbookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get

func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := logf.FromContext(ctx)
    logger.Info("Starting reconciliation", "namespace", req.Namespace, "name", req.Name)

    // 1. Fetch Guestbook CR
    logger.Info("Fetching Guestbook CR")
    var guestbook webappv1.Guestbook
    if err := r.Get(ctx, req.NamespacedName, &guestbook); err != nil {
        if errors.IsNotFound(err) {
            logger.Info("Guestbook not found, ignoring")
            return ctrl.Result{}, nil
        }
        logger.Error(err, "Failed to fetch Guestbook")
        return ctrl.Result{}, err
    }
    logger.Info("Guestbook fetched", "spec", guestbook.Spec)

    // 2. Ensure ConfigMap exists
    logger.Info("Checking ConfigMap", "name", guestbook.Spec.ConfigMapName)
    var configContent string
    switch guestbook.Spec.Type {
    case "Phone":
        configContent = "type: Phone\nnumber: 123-456-7890"
    case "Address":
        configContent = "type: Address\nstreet: 123 Main St\ncity: Dhaka"
    case "Name":
        configContent = "type: Name\nfirst: John\nlast: Doe"
    default:
        configContent = "type: Unknown"
    }

    cm := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      guestbook.Spec.ConfigMapName,
            Namespace: req.Namespace,
            Labels: map[string]string{
                "app": guestbook.Name,
            },
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(&guestbook, webappv1.GroupVersion.WithKind("Guestbook")),
            },
        },
        Data: map[string]string{
            "config.yaml": configContent,
        },
    }
    foundCM := &corev1.ConfigMap{}
    err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: cm.Name}, foundCM)
    if err != nil {
        if errors.IsNotFound(err) {
            logger.Info("Creating ConfigMap", "name", cm.Name)
            if err := r.Create(ctx, cm); err != nil {
                logger.Error(err, "Failed to create ConfigMap")
                return ctrl.Result{}, err
            }
            logger.Info("ConfigMap created", "name", cm.Name)
        } else {
            logger.Error(err, "Failed to get ConfigMap")
            return ctrl.Result{}, err
        }
    } else {
        logger.Info("Updating ConfigMap", "name", foundCM.Name)
        foundCM.Data = cm.Data
        if err := r.Update(ctx, foundCM); err != nil {
            logger.Error(err, "Failed to update ConfigMap")
            return ctrl.Result{}, err
        }
        logger.Info("ConfigMap updated", "name", foundCM.Name)
    }

    // 3. List existing pods
    logger.Info("Listing pods with label", "app", guestbook.Name)
    podList := &corev1.PodList{}
    listOpts := []client.ListOption{
        client.InNamespace(req.Namespace),
        client.MatchingLabels{"app": guestbook.Name},
    }
    if err := r.List(ctx, podList, listOpts...); err != nil {
        logger.Error(err, "Failed to list pods")
        return ctrl.Result{}, err
    }
    logger.Info("Found pods", "count", len(podList.Items))

    // 4. Create pods if needed
    currentPods := len(podList.Items)
    desiredPods := int(guestbook.Spec.Size)
    logger.Info("Checking pod count", "current", currentPods, "desired", desiredPods)

    // Build dynamic container args
    authArg := "--auth=false"
    if guestbook.Spec.Auth != nil && *guestbook.Spec.Auth {
        authArg = "--auth=true"
    }
    portArg := fmt.Sprintf("--port=%d", guestbook.Spec.Port)

    for i := currentPods; i < desiredPods; i++ {
        logger.Info("Creating pod", "index", i)
        pod := &corev1.Pod{
            ObjectMeta: metav1.ObjectMeta{
                Name:      fmt.Sprintf("%s-pod-%d", guestbook.Name, i),
                Namespace: req.Namespace,
                Labels:    map[string]string{"app": guestbook.Name},
                OwnerReferences: []metav1.OwnerReference{
                    *metav1.NewControllerRef(&guestbook, webappv1.GroupVersion.WithKind("Guestbook")),
                },
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{{
                    Name:    "guestbook",
                    Image:   guestbook.Spec.Image,
                    Command: []string{"./main", "startProject"},
                    Args:    []string{portArg, authArg},
                    Env: []corev1.EnvVar{
                        {
                            Name: "JWT_SECRET",
                            ValueFrom: &corev1.EnvVarSource{
                                SecretKeyRef: &corev1.SecretKeySelector{
                                    LocalObjectReference: corev1.LocalObjectReference{
                                        Name: guestbook.Spec.JWTSecretName,
                                    },
                                    Key: "JWT_SECRET",
                                },
                            },
                        },
                    },
                    VolumeMounts: []corev1.VolumeMount{
                        {
                            Name:      "guestbook-config",
                            MountPath: "/etc/guestbook",
                        },
                    },
                }},
                Volumes: []corev1.Volume{
                    {
                        Name: "guestbook-config",
                        VolumeSource: corev1.VolumeSource{
                            ConfigMap: &corev1.ConfigMapVolumeSource{
                                LocalObjectReference: corev1.LocalObjectReference{
                                    Name: guestbook.Spec.ConfigMapName,
                                },
                            },
                        },
                    },
                },
            },
        }
        if err := r.Create(ctx, pod); err != nil {
            logger.Error(err, "Failed to create pod", "name", pod.Name)
            return ctrl.Result{}, err
        }
        logger.Info("Created pod", "name", pod.Name)
    }

    // 5. Delete extra pods
    logger.Info("Checking for extra pods")
    for i := desiredPods; i < currentPods; i++ {
        podName := fmt.Sprintf("%s-pod-%d", guestbook.Name, i)
        pod := &corev1.Pod{}
        err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: podName}, pod)
        if err != nil {
            if errors.IsNotFound(err) {
                logger.Info("Pod not found, skipping deletion", "name", podName)
                continue
            }
            logger.Error(err, "Failed to get pod for deletion", "name", podName)
            return ctrl.Result{}, err
        }
        logger.Info("Deleting pod", "name", podName)
        if err := r.Delete(ctx, pod); err != nil {
            logger.Error(err, "Failed to delete pod", "name", podName)
            return ctrl.Result{}, err
        }
        logger.Info("Deleted pod", "name", podName)
    }

    // 6. Update Guestbook status with retry on conflict
    logger.Info("Updating Guestbook status")
    var active string
    var standby []string
    for i, pod := range podList.Items {
        if i == 0 {
            active = pod.Name
        } else {
            standby = append(standby, pod.Name)
        }
    }

    // Retry status update on conflict
    err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
        // Fetch the latest version of the Guestbook
        latestGuestbook := &webappv1.Guestbook{}
        if err := r.Get(ctx, req.NamespacedName, latestGuestbook); err != nil {
            return err
        }
        // Update status fields
        latestGuestbook.Status.Active = active
        latestGuestbook.Status.Standby = standby
        // Attempt to update status
        return r.Status().Update(ctx, latestGuestbook)
    })
    if err != nil {
        logger.Error(err, "Failed to update Guestbook status after retries")
        return ctrl.Result{}, err
    }
    logger.Info("Guestbook status updated", "active", active, "standby", standby)

    logger.Info("Reconciliation complete")
    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Guestbook{}).
		Named("guestbook").
		Complete(r)
}
