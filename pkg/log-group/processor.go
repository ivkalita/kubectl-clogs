package log_group

import (
	"bufio"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type logLine struct {
	line string
	pod  *v1.Pod
}

type processor struct {
	pod       *v1.Pod
	clientSet *kubernetes.Clientset
}

func makeProcessor(pod *v1.Pod, clientSet *kubernetes.Clientset) processor {
	t := processor{pod, clientSet}
	return t
}

func (p processor) process(out chan logLine) {
	tailLines := int64(10)
	options := &v1.PodLogOptions{Follow: true, TailLines: &tailLines}
	rc, err := p.clientSet.CoreV1().Pods(p.pod.Namespace).GetLogs(p.pod.Name, options).Stream()
	if err != nil {
		panic(err)
	}

	defer rc.Close()

	r := bufio.NewReader(rc)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		out <- logLine{line, p.pod}
	}
}
