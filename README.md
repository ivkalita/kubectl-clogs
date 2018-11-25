## kubectl-clogs

This repository contains a simple kubectl command to show all running pods logs in specific cluster namespace in realtime
(the same as `kubectl logs -f` but for the all pods).


```bash
$ go get github.com/kaduev13/kubectl-clogs/cmd/kubectl-clogs
$ kubectl clogs --namespace default
```

![](https://i.imgur.com/zG8eG3k.jpg)


## Features:
- Logs from all namespace pods in realtime (follow mode)
- Autoreload – this command monitors all running pods every `N` seconds and adds new pods to the log group

## Roadmap:
- Regexp for the pod name
- Multiple namespaces
