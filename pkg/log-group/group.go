package log_group

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"time"
)

type colorizeFn func(interface{}) aurora.Value

type LogGroup struct {
	clientSet *kubernetes.Clientset
	namespace string
	colorMap  map[types.UID]colorizeFn
}

var colorizers = [...]colorizeFn{aurora.Green, aurora.Magenta, aurora.Cyan, aurora.Red, aurora.Brown, aurora.Blue}

func New(clientSet *kubernetes.Clientset, namespace string) LogGroup {
	return LogGroup{clientSet, namespace, make(map[types.UID]colorizeFn)}
}

func (lg LogGroup) Tail() {
	logLineChan := make(chan logLine)

	go func() {
		lg.monitorPods(logLineChan)
	}()

	var lastPod *v1.Pod = nil
	for logLine := range logLineChan {
		colorize := lg.colorMap[logLine.pod.UID]
		if lastPod == nil || logLine.pod.UID != lastPod.UID {
			fmt.Println(aurora.Inverse(colorize(fmt.Sprintf("[%s]", logLine.pod.Name))))
			lastPod = logLine.pod
		}
		fmt.Print(colorize(logLine.line))
	}
}

func (lg LogGroup) allPods() *v1.PodList {
	pods, err := lg.clientSet.CoreV1().Pods(lg.namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return pods
}

func (lg LogGroup) monitorPods(logLineChan chan logLine) {
	for {
		pods := lg.allPods()
		for i, p := range pods.Items {
			if p.Status.Phase != v1.PodRunning {
				continue
			}
			if _, ok := lg.colorMap[p.UID]; ok {
				continue
			}
			lg.colorMap[p.UID] = colorizers[i%len(colorizers)]
			fmt.Printf("New pod %s found\n", aurora.Inverse(lg.colorMap[p.UID](p.Name)))
			go func(p v1.Pod) {
				proc := makeProcessor(&p, lg.clientSet)
				proc.process(logLineChan)
			}(p)
		}

		time.Sleep(10 * time.Second)
	}
}
