package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ivkalita/kubectl-clogs/internal/streams"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
)

type args struct {
	kubeConfig    string
	namespace     string
	podNameRegexp string
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	args := parseArguments()
	cs, err := createClientSet(args.kubeConfig)
	if err != nil {
		return fmt.Errorf("prepare k8s client: %w", err)
	}

	group := streams.NewGroup(cs, args.namespace, regexp.MustCompile(args.podNameRegexp))

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer cancel()
		return group.Tail(ctx)
	})
	eg.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case <-c:
			cancel()
		}
		return nil
	})

	return eg.Wait()
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

func createClientSet(kubeConfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("build config from flags (%s): %w", kubeConfig, err)
	}

	cs, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, fmt.Errorf("new for config (%s): %w", kubeConfig, err)
	}

	return cs, nil
}
