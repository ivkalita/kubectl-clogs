package streams

import (
	"bufio"
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type logLine struct {
	line        string
	podUID      types.UID
	containerID string
}

type processor struct {
	podName       string
	podUID        types.UID
	podNamespace  string
	containerName string
	containerID   string
	client        *kubernetes.Clientset
}

func newProcessor(
	podUID types.UID, podName string, podNamespace string,
	containerName string, containerID string,
	client *kubernetes.Clientset,
) processor {
	t := processor{
		podName:       podName,
		podUID:        podUID,
		podNamespace:  podNamespace,
		containerName: containerName,
		containerID:   containerID,
		client:        client,
	}
	return t
}

func (p processor) process(ctx context.Context, out chan logLine) error {
	tailLines := int64(10)
	options := &v1.PodLogOptions{Follow: true, TailLines: &tailLines, Container: p.containerName}
	rc, err := p.client.CoreV1().Pods(p.podNamespace).GetLogs(p.podName, options).Stream(ctx)
	if err != nil {
		return fmt.Errorf("stream (%s – %s): %w", p.podName, p.containerName, err)
	}

	defer rc.Close()

	r := bufio.NewReader(rc)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		select {
		case out <- logLine{
			line:        line,
			podUID:      p.podUID,
			containerID: p.containerID,
		}:
			break
		case <-ctx.Done():
			return nil
		}
	}

	return nil
}
