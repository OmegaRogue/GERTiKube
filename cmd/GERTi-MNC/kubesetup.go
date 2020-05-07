package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func InternalKube() (kubernetes.Clientset, error) {
	if h := os.Getenv("KUBERNETES_SERVICE_HOST"); h != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return kubernetes.Clientset{}, fmt.Errorf("error on rest.InClusterConfig: %w", err)
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return kubernetes.Clientset{}, fmt.Errorf("error on kubernetes.NewForConfig: %w", err)
		}
		return *clientset, nil
	}
	return kubernetes.Clientset{}, errors.New("error accessing kubernetes cluster: not inside kubernetes cluster")

}

func ExternalKube() (kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return kubernetes.Clientset{}, fmt.Errorf("error on BuildConfigFromFlags: %w", err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return kubernetes.Clientset{}, fmt.Errorf("error on kubernetes.NewForConfig: %w", err)
	}

	return *clientset, nil
}

func ClusterSetup() (kubernetes.Clientset, error) {
	clientset, err := InternalKube()
	if err != nil {
		log.Printf("error on InternalKube: %+v\n", err)
		clientset, err := ExternalKube()
		if err != nil {
			return kubernetes.Clientset{}, fmt.Errorf("error on ExternalKube: %w", err)
		}
		return clientset, nil
	}
	return clientset, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
