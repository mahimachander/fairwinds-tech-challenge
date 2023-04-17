package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	namespaceAnnotationKey = "managed"
	podAnnotationKey       = "managed"

	leaseName      = "pod-controller-lease"
	leaseNamespace = "default"
)

func main() {
	// Initialize Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	leaderElectionConfig := leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      leaseName,
				Namespace: leaseNamespace,
			},
			Client: clientset.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				// Unique identifier for this process across all participants in the election
				Identity: string(uuid.NewUUID()),
			},
		},
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// Start processing tasks as the leader
				watchForNewPods(clientset)
			},
			OnStoppedLeading: func() {
				// Stop processing tasks when leadership is lost
			},
			OnNewLeader: func(identity string) {
				// New leader has been elected
			},
		},
	}

	leaderelection.RunOrDie(context.Background(), leaderElectionConfig)
}

// TODO: exclude existing pods, ResourceVersion: "" as a ListOption doesn't work
func watchForNewPods(clientset *kubernetes.Clientset) {
	podWatch, err := clientset.CoreV1().Pods(metav1.NamespaceAll).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	defer podWatch.Stop()

	// Process events received from pod watch
	for event := range podWatch.ResultChan() {
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		if event.Type == watch.Added && validatePod(clientset, pod) {
			annotatePod(clientset, pod)
			logPod(pod)
		}
	}
}

// Check if a pod has a specific annotation and if its namespace has a specific annotation
func validatePod(clientset *kubernetes.Clientset, pod *corev1.Pod) bool {
	ns, err := clientset.CoreV1().Namespaces().Get(context.Background(), pod.GetNamespace(), metav1.GetOptions{})
	if err != nil {
		fmt.Println(err)
	}
	nsAnnotations := ns.GetAnnotations()
	if nsAnnotations == nil {
		return false
	}

	podAnnotations := pod.GetAnnotations()
	if podAnnotations == nil {
		return false
	}

	return pod.GetAnnotations()[podAnnotationKey] != "" && ns.GetAnnotations()[namespaceAnnotationKey] != ""
}

// Annotate pod with timestamp
func annotatePod(clientset *kubernetes.Clientset, pod *corev1.Pod) {
	timestamp := time.Now().Format(time.RFC3339)
	annotations := pod.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["timestamp"] = timestamp
	pod.SetAnnotations(annotations)

	_, err := clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println(err)
	}
}

// Log pod name and timestamp to stdout
func logPod(pod *corev1.Pod) {
	timestamp := pod.GetAnnotations()["timestamp"]
	fmt.Printf("New Pod: %s (timestamp: %s)\n", pod.Name, timestamp)
}
