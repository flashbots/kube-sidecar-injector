# kube-sidecar-injector

Initial implementation of the sidecar injector for k8s.

## TL;DR

With configuration like this `kube-sidecar-injector` will make sure that any
container that runs in EKS fargate will have prometheus node-exporter sidecar
running next to it.

```yaml
inject:
  - labelSelector:
      matchExpressions:
        - key: eks.amazonaws.com/fargate-profile
          operator: Exists

    labels:
      flashbots.net/prometheus-node-exporter: true

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

```

### Caveats

Single webhook configuration can me configured to apply multiple injection
rules.  However, if these rules are supposed to interact somehow (for example
rule A introduces changes that rule B is supposed to act upon) then they should
be placed into _separate_ webhooks.

See k8s webhook [reinvocation policy](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#reinvocation-policy)
for the details.
