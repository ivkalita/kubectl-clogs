package streams

import (
	"context"
	"fmt"
	"github.com/logrusorgru/aurora"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"regexp"
	"time"
)

type group struct {
	client        *kubernetes.Clientset
	namespace     string
	podNameRegexp *regexp.Regexp
	pods          map[types.UID]pod
}

type pod struct {
	name       string
	containers map[string]container
}

type container struct {
	name     string
	colorize colorizeFn
}

type colorizeFn func(interface{}) aurora.Value

var colorizers = [...]colorizeFn{aurora.Green, aurora.Magenta, aurora.Cyan, aurora.Red, aurora.Brown, aurora.Blue}

func NewGroup(client *kubernetes.Clientset, namespace string, podNameRegexp *regexp.Regexp) *group {
	return &group{
		client:        client,
		namespace:     namespace,
		podNameRegexp: podNameRegexp,
		pods:          make(map[types.UID]pod),
	}
}

func (g *group) Tail(ctx context.Context) error {
	logLineChan := make(chan logLine)
	defer close(logLineChan)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		const updateInterval = 10 * time.Second
		t := time.NewTimer(updateInterval)
		defer t.Stop()

		for {
			err := g.discoverPods(ctx, eg, logLineChan)
			if err != nil {
				return fmt.Errorf("discover pods: %w", err)
			}
			t.Reset(updateInterval)
			select {
			case <-ctx.Done():
				return nil
			case <-t.C:
			}
		}
	})
	eg.Go(func() error {
		var (
			lastPodUID      types.UID
			lastContainerID string
		)
		for logLine := range logLineChan {
			p := g.pods[logLine.podUID]
			c := p.containers[logLine.containerID]
			if logLine.podUID != lastPodUID && logLine.containerID != lastContainerID {
				fmt.Println(aurora.Inverse(c.colorize(fmt.Sprintf("[%s – %s]", p.name, c.name))))
				lastPodUID = logLine.podUID
				lastContainerID = logLine.containerID
			}
			fmt.Print(c.colorize(logLine.line))
		}
		return nil
	})

	return eg.Wait()
}

func (g *group) allPods(ctx context.Context) (*v1.PodList, error) {
	pods, err := g.client.CoreV1().Pods(g.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	return pods, nil
}

func (g *group) discoverPods(ctx context.Context, eg *errgroup.Group, logLineChan chan logLine) error {
	pods, err := g.allPods(ctx)
	if err != nil {
		return fmt.Errorf("all pods: %w", err)
	}
	for i, p := range pods.Items {
		if !g.podNameRegexp.MatchString(p.Name) {
			// pod is filtered
			continue
		}
		for j, c := range p.Status.ContainerStatuses {
			pp, ok := g.pods[p.UID]
			if !ok {
				pp = pod{
					name:       p.Name,
					containers: make(map[string]container),
				}
			}
			_, ok = pp.containers[c.ContainerID]
			if ok {
				// container is already known, skipping
				continue
			}
			// container is not known yet
			cc := container{
				name:     c.Name,
				colorize: colorizers[(i*5+j)%len(colorizers)],
			}
			pp.containers[c.ContainerID] = cc
			g.pods[p.UID] = pp

			func(pUID types.UID, pName string, pNamespace string, cName string, cID string) {
				eg.Go(func() error {
					proc := newProcessor(pUID, pName, pNamespace, cName, cID, g.client)
					return proc.process(ctx, logLineChan)
				})
			}(p.UID, p.Name, p.Namespace, c.Name, c.ContainerID)
		}
	}

	return nil
}
