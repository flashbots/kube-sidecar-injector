# kube-sidecar-injector

Initial implementation of the sidecar injector for k8s.

## TL;DR

With configuration like this `kube-sidecar-injector` will make sure that any
container that runs in EKS fargate will have prometheus node-exporter sidecar
running next to it.

```yaml
inject:
  - labels:
      flashbots.net/fargate-node-exporter: true

    containers:
      - name: node-exporter
        image: prom/node-exporter:v1.7.0
        args: [
          "--log.format", "json",
          "--web.listen-address", ":9100",
        ]
        ports:
          - name: http-metrics
            containerPort: 9100
        resources:
          requests:
            cpu: 10m
            memory: 64Mi

    labelSelector:
      matchExpressions:
        - key: eks.amazonaws.com/fargate-profile
          operator: Exists
```
