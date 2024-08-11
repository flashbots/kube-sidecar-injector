# kube-sidecar-injector

Initial implementation of the sidecar injector for k8s.

## TL;DR

1.  With configuration like this `kube-sidecar-injector` will make sure that any
    container that runs in EKS fargate will have prometheus node-exporter sidecar
    running next to it:

    ```yaml
    inject:
      - name: inject-node-exporter

        labelSelector:
          matchExpressions:
            - key: eks.amazonaws.com/fargate-profile
              operator: Exists

        namespaceSelector:
          matchExpressions:
            - key: kubernetes.io/metadata.name
              operator: NotIn
              values: [kube-system]

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
              - name: node-exporter
                containerPort: 9100
            resources:
              requests:
                cpu: 10m
                memory: 64Mi
    ```

2.  In conjunction with `trust-manager` this will allow to automatically mount
    root CA in every pod:

    ```yaml
    inject:
      - name: inject-internal-ca

        volumes:
          - name: internal-ca
            configMap:
              name: internal-ca

        volumeMounts:
          - mountPath: /usr/local/share/ca-certificates
            name: internal-ca
            readOnly: true

          - mountPath: /etc/ssl/certs/internal-ca.crt
            name: internal-ca
            subPath: internal-ca.crt
            readOnly: true
    ```

### Caveats

- Single webhook configuration can be configured to apply multiple injection
  rules.  However, if these rules should interact somehow (for example rule A
  introduces changes that rule B is supposed to act upon) then these rules
  should be placed into _separate_ webhooks.

  See k8s webhook [reinvocation policy](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#reinvocation-policy)
  for the details.

- It's not possible for the webhook to know at the runtime whether the patch it
  generates is invalid.

  For example, if you try to inject a container that has port name of more than
  15 characters long k8s will not allow the modified pod to be deployed.

  In situations like this, k8s will infinitely attempt the webhook admission,
  without ever creating the pod.  In order to troubleshoot this issue it could
  help to see actual underlying error from k8s with:

  ```shell
  kubectl get events
  ```
