package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	// Watch for new pods
	podWatch, err := clientset.CoreV1().Pods(v1.NamespaceAll).Watch(context.Background(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	defer podWatch.Stop()

	// Process events received from pod watch
	// TODO: Only respond to pods with a particular annotation
	// TODO: Only respond to pods in namespaces with a particular annotation
	for event := range podWatch.ResultChan() {
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		switch event.Type {
		case watch.Added:
			annotatePod(clientset, pod)
			logPod(pod)
		}
	}
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

	_, err := clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, v1.UpdateOptions{})
	if err != nil {
		fmt.Println(err)
	}
}

// Log pod name and timestamp to stdout
func logPod(pod *corev1.Pod) {
	timestamp := pod.GetAnnotations()["timestamp"]
	fmt.Printf("New Pod: %s (timestamp: %s)\n", pod.Name, timestamp)
}
