package main

import (
	"flag"
	"github.com/kaduev13/kubectl-clogs/pkg/log-group"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type args struct {
	kubeConfig    string
	namespace     string
	podNameRegexp string
}

func main() {
	args := parseArguments()
	clientset := createClientSet(args.kubeConfig)

	g := log_group.New(clientset, args.namespace, args.podNameRegexp)
	g.Tail()
}

func parseArguments() args {
	defaultPath := ""
	if home := os.Getenv("HOME"); home != "" {
		defaultPath = filepath.Join(home, ".kube", "config")
	}
	kubeConfig := flag.String("kubeconfig", defaultPath, "path to the kubeconfig file")
	namespace := flag.String("namespace", "default", "namespace")
	podNameRegexp := ".*"
	flag.Parse()
	if flag.NArg() > 0 {
		podNameRegexp = flag.Arg(0)
	}

	return args{*kubeConfig, *namespace, podNameRegexp}
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
