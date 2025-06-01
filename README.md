# glance-k8s

An extension widget for the amazing [Glance][1] dashboard.

## Features

This extension features two widgets to integrate Kubernetes resources into your dashboard.

### Kubernetes Nodes

![Kubernetes Nodes](docs/images/nodes.png)

#### Setup

```yaml
widgets:
  - type: extension
    url: http://glance-k8s/extension/nodes
    allow-potentially-dangerous-html: true
    cache: 1s
```

### Kubernetes Applications

![Kubernetes Apps](docs/images/apps.png)

#### Setup

```yaml
widgets:
  - type: extension
    url: http://glance-k8s/extension/apps
    allow-potentially-dangerous-html: true
    cache: 1s

    parameters:
      # Parameters are sent to the extension as query parameters.
      # Since there is no browser or proxy between glance and glance-k8s there _should_ not be a size limit for the parameters,
      # apart from default buffer sizes.
      # But keep that in mind, when supplying large strings.

      # Show only workloads, that match an expression. 
      # See: https://github.com/expr-lang/expr
      #
      # Environment:
      #   namespace    Namespace of the workload
      #   name         Name of the workload
      #   annotations  Map of annotations
      show-if: |
        namespace != "kube-system" and
        ("glance/hide" not in annotations || annotations["glance/hide"] != "true")

      # You can also supply multiple expressions, which will evaluate to a logical conjunction (AND).
      # show-if:
      #   - |
      #     namespace != "kube-system"
      #   - |
      #     ("glance/hide" not in annotations || annotations["glance/hide"] != "true")
```

#### Customization / How it works

Kubernetes has a lot of moving parts, which makes it a little tricky to find all the installed applications.

The extension iterates over workloads (`Deployment`, `StatefulSet` and `DaemonSet`), services and ingresses in all namespaces.

Then it tries to match workloads to services and services to ingresses using their specified selectors to find an ingress for applications.
Since configurations can become very complex, it might not be able to find the right ingress, if more than one exists.
For most cases however, it should just work.

Finally the workloads are grouped into applications, which belong together. 
If you do not annotate workloads, every workload is assumed to be an application.

You can annotate workloads to group them into applications and customize their appearance on the dashboard.
If the workload has an ingress, you may annotate the ingress as well.

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    # Display Name (default: TitleCase(workload.Name))
    glance/name: Glance

    # Icon (default: di:kubernetes)
    # The shorthands `di:` and `si:` are supported similar to glance.
    glance/icon: di:glance
    # glance/icon: https://example.org/glance.png

    # Link to the application (default: Ingress with shortest path. First found in order: Main Workload > Dependencies)
    glance/url: https://glance.example.org

    # Open links on the same tab (default: false)
    glance/same-tab: true

    # Description
    glance/description: My fancy dashboard

    # Identifier for an application to group workloads.
    # This should be annotated on the "main" workload of an application.
    glance/id: glance

    # Identifier of the main workload of the same app.
    # This should be annotated on all workloads other than the "main" one of the same application.
    # If multiple workloads have the same parent but there is no workload annotated with `glance/id`
    # the first found will be promoted to be the "main" workload.
    glance/parent: glance
```


## Installation

Glance itself provides a container image, but no official helm chart yet.
Until then, this repository contains charts for both upstream glance as well as the glance-k8s extension.

Updates of glance are tracked using a self-hosted [Renovate Bot][2] running on github actions.

```bash
# Use the latest versions instead
export GLANCE_VERSION=v0.8.3
export GLANCE_K8S_VERSION=v0.1.3

# See https://helm.sh/docs/helm/helm_install/
# You can provide a values file to the install command using `-f values.yaml`
helm install glance oci://ghcr.io/lukasdietrich/glance-k8s/chart/glance:${GLANCE_VERSION}
helm install glance-k8s oci://ghcr.io/lukasdietrich/glance-k8s/chart/glance-k8s:${GLANCE_K8S_VERSION}
```
### Values

The default values can be found in their respective chart folder:
- [glance](charts/glance/values.yaml)
- [glance-k8s](charts/glance-k8s/values.yaml)

[1]: https://github.com/glanceapp/glance
[2]: https://github.com/renovatebot/renovate
