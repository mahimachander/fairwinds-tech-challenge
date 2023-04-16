package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const(
	namespaceAnnotationKey = "managed"
	podAnnotationKey
)

func main() {
	// Initialize Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Watch for new pods. Exclude existing pods
	podWatch, err := clientset.CoreV1().Pods(metav1.NamespaceAll).Watch(context.Background(), metav1.ListOptions{ResourceVersion: ""})
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
