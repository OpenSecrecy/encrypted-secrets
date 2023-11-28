package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func MdHashing(input string) string {
	byteInput := []byte(input)
	md5Hash := md5.Sum(byteInput)
	return hex.EncodeToString(md5Hash[:]) // by referring to it as a string
}

func GetKubeClient() (*kubernetes.Clientset, error) {

	// it running inside a k8s cluster, use InCLusterConfig
	if runningInKubernetes() {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build client configuration: %v", err)
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
		}
		return clientset, nil
	}

	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// check for KUBECONFIG env variable
	if os.Getenv("KUBECONFIG") != "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	// Build the client configuration from the provided kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build client configuration: %v", err)
	}

	// Create a new Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	return clientset, err

}

func runningInKubernetes() bool {

	// Check for the presence of common Kubernetes environment variables
	envVars := []string{"KUBERNETES_SERVICE_HOST", "KUBERNETES_PORT"}
	for _, envVar := range envVars {
		if _, exists := os.LookupEnv(envVar); !exists {
			return false
		}
	}

	return true
}
