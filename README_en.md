# Astra-xmod-shim

A lightweight middleware for unified AI model serving orchestration.

## Overview

Astra-xmod-shim decouples *where* a model runs from *how* it is deployed. It uses **Shimlets** to abstract runtime environments and **Pipelines** to define deployment workflows—enabling consistent management across platforms.

Designed for extensibility and minimal footprint, it runs as a single binary with no external dependencies.

---

## Architecture
![架构示意图](img.png)

- **Core Engine**: Manages service lifecycle via a finite state machine (FSM), handles API requests, and coordinates plugins.
- **Shimlet**: Adapts to runtime environments (e.g., Kubernetes, Docker) through a plugin interface.
- **Pipeline**: Composes deployment logic as a chain of function steps.
- **EventBus**: Decouples core operations from observability and extension systems.

```
  API/CLI
    │
    ▼
  Core Engine (FSM)
    ├─▶ Shimlet (Runtime)
    ├─▶ Pipeline (Workflow)
    └─▶ EventBus → Logging, Monitoring, etc.
```

---

## Quick Start

```bash
# Download and run
wget https://github.com/iflytek/modserv-shim/releases/latest/download/model-serve-shim
chmod +x model-serve-shim

./model-serve-shim \
  --port=8080 \
  --shimlet=k8s \
  --pipeline=opensourcellm
```

## API Example

Deploy a model:
```bash
curl -X POST http://localhost:8080/api/v1/modserv/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "modelName": "qwen",
    "resourceRequirements": {
      "acceleratorType": "NVIDIA GPU",
      "acceleratorCount": 1,
      "cpu": "4",
      "memory": "16Gi"
    },
    "replicaCount": 1
  }'
```

Check status:
```bash
curl http://localhost:8080/api/v1/modserv/{serviceId}
```

---

## Example: OpenSourceLLM Pipeline

A built-in pipeline for deploying open-source LLMs. Uses a builder pattern to define ordered steps.

```go
// mypipeline/mypipeline.go
package mypipeline

import (
  "modserv-shim/internal/core/pipeline"
  "modserv-shim/pkg/log"
)

func validate(ctx *pipeline.Context) error {
  log.Info("Validating model: %s", ctx.DeploySpec.ModelName)
  // Add model path, format checks
  return nil
}

func generateConfig(ctx *pipeline.Context) error {
  log.Info("Generating runtime config")
  // Set up inference server args, env vars
  return nil
}

func exposeService(ctx *pipeline.Context) error {
  log.Info("Exposing service endpoint")
  // Create service ingress/route
  return nil
}

// Register the pipeline
func init() {
  pipeline.New("opensourcellm").
    Step(validate).
    Step(generateConfig).
    Step(exposeService).
    BuildAndRegister()
}
```

> This pipeline can be selected at startup: `--pipeline=opensourcellm`.

---

## Example: Kubernetes Shimlet

A built-in runtime adapter that deploys models on Kubernetes using native APIs.

```go
// k8s/shimlet.go
package k8s

import (
  "modserv-shim/internal/core/deploy"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sShimlet struct {
  client *kubernetes.Clientset
}

func (k *K8sShimlet) Create(ctx *deploy.Context) (string, error) {
  // Convert deploy spec to Kubernetes Deployment + Service
  deployment := &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{Name: ctx.ServiceID},
    // ... setup containers, resources, replicas
  }
  _, err := k.client.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
  if err != nil {
    return "", err
  }
  return ctx.ServiceID, nil
}

func (k *K8sShimlet) Status(resourceID string) (deploy.Status, error) {
  // Query pod status, return Running/Failed/Pending
  pods, err := k.client.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
    LabelSelector: "service-id=" + resourceID,
  })
  // Analyze pod states
  return deploy.StatusRunning, nil
}

func (k *K8sShimlet) Delete(resourceID string) error {
  // Delete deployment and service
  return k.client.AppsV1().Deployments("default").Delete(context.TODO(), resourceID, metav1.DeleteOptions{})
}

// Register at init
func init() {
  plugin.RegisterShimlet("k8s", &K8sShimlet{client: getK8sClient()})
}
```

> Enabled via: `--shimlet=k8s`.

---

## Extensibility

### Custom Use Cases

- **Edge Pipeline**: Add model quantization, offline support, and resource throttling.
- **Multimodal Pipeline**: Extend validation for image/text inputs and GPU memory tuning.
- **Enterprise Pipeline**: Inject auth, encryption, and audit logging.
- **Docker Shimlet**: Target container runtimes without Kubernetes.

Plugins are compiled into the binary via Go’s `init()` registration mechanism.

---

## Configuration

Via flags:
```bash
--port=8080 --shimlet=k8s --pipeline=opensourcellm --log-level=info
```

Or YAML:
```yaml
service:
  port: 8080
plugins:
  defaultShimlet: k8s
  defaultPipeline: opensourcellm
logging:
  level: info
```

---

## License

Apache License 2.0

## Contact

- Issues: [GitHub Issues](https://github.com/iflytek/modserv-shim/issues)
- Email: hxli28@iflytek.com
