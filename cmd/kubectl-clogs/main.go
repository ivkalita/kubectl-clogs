package main

import (
	"flag"
	"github.com/kaduev13/kubectl-clogs/pkg/log-group"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func main() {
	kubeConfig, namespace := parseArguments()
	clientset := createClientSet(*kubeConfig)

	g := log_group.New(clientset, *namespace)
	g.Tail()
}

func parseArguments() (*string, *string) {
	defaultPath := ""
	if home := os.Getenv("HOME"); home != "" {
		defaultPath = filepath.Join(home, ".kube", "config")
	}
	kubeConfig := flag.String("kubeconfig", defaultPath, "path to the kubeconfig file")
	namespace := flag.String("namespace", "default", "namespace")

	flag.Parse()

	return kubeConfig, namespace
}

func createClientSet(kubeConfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	return clientset
}
