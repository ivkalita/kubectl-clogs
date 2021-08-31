## kubectl-clogs

This repository contains a simple kubectl command to show all running pods logs in specific cluster namespace in realtime
(the same as `kubectl logs -f` but for the all pods).


```bash
$ go install github.com/ivkalita/kubectl-clogs/cmd/kubectl-clogs@latest
$ kubectl clogs --namespace default
```

![](https://i.imgur.com/zG8eG3k.jpg)


## Features:
- Logs from all namespace pods in realtime (follow mode)
- Autoreload – this command monitors all running pods every `N` seconds and adds new pods to the log group
- Regexp for the pod name

## Roadmap:
- Multiple namespaces
- Labels search
- Monitoring improvements (`watch` mode)
