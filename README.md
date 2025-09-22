## Project Overview
ModelServeShim is a lightweight AI service orchestration middleware designed with a plugin-based architecture to simplify the deployment, operation, and management of large language model services. By abstracting the environment adaptation layer (shimlet) and extendable deployment workflows (pipeline), it achieves unified cross-environment management capabilities and supports rapid integration of new deployment environments and custom model deployment workflows.

## Core Features
- **Plugin-based Cross-environment Adaptation**: Based on the shim abstraction layer design, enables seamless adaptation to different environments through shimlet plugins
- **Customizable Deployment Workflows**: Based on the pipeline plugin architecture, supports custom model deployment steps and workflows
- **Lightweight and Efficient**: Single binary delivery with no external dependencies and low resource consumption
- **Full Lifecycle Management**: Supports complete lifecycle management of model services including deployment, monitoring, updating, and destruction
- **Hot Reload Mechanism**: Supports hot reloading of plugins, enabling feature updates without restarting the service

## Technical Architecture
ModelServeShim adopts a "core logic + dual plugins" decoupled design architecture, mainly consisting of the following components:

1. **Core Engine**: Responsible for overall process coordination, state management, and API provision
2. **Shim Abstraction Layer**: Defines standardized interfaces to shield underlying environment differences
3. **Pipeline Engine**: Manages various stages and steps of model deployment
4. **Plugin Management System**: Handles loading, unloading, and lifecycle management of plugins

![Architecture Diagram]()

### Core Component Description
- **Core Engine**: Processes API requests, manages service states, and coordinates the work of various components
- **Shim Layer**: Implements adaptation to different environments (such as K8s, Docker) through unified interface definitions
- **Pipeline Layer**: Defines and executes various steps of model deployment, such as model validation, configuration rendering, resource deployment, etc.
- **State Management**: Implements reliable service state transitions and tracking based on finite state machines

## Quick Start
### Environment Requirements
- Go 1.20+ (development environment)
- Target environment (e.g., K8s v1.20+, if using K8s shimlet)

### Installation
```bash
# Download binary file (Linux x86_64)
wget https://github.com/iflytek/modserv-shim/releases/latest/download/model-serve-shim
chmod +x model-serve-shim

# Or build from source
git clone https://github.com/iflytek/modserv-shim.git
cd modserv-shim
make build
```

### Basic Usage
```bash
# Start the service, loading K8s shimlet and open source LLM deployment workflow
./model-serve-shim --port=8080 \
  --shimlet=k8s \
  --pipeline=opensourcellm
```

## API Reference
### Deploy Model Service
```bash
curl -X POST http://localhost:8080/api/v1/modserv/deploy \
  -H "Content-Type: application/json" \
  -d '{ \
    "modelName": "example-model", \
    "modelFile": "/path/to/model", \
    "resourceRequirements": { \
      "acceleratorType": "NVIDIA GPU", \
      "acceleratorCount": 1, \
      "cpu": "4", \
      "memory": "16Gi" \
    }, \
    "replicaCount": 1 \
  }'
```

### Query Service Status
```bash
curl http://localhost:8080/api/v1/modserv/{serviceId}
```

### List Loaded Plugins
```bash
curl http://localhost:8080/api/v1/plugins
```

## Plugin Development Guide
### Shimlet Development (Environment Adaptation Plugin)
Shimlet is responsible for converting abstract deployment requests into operations specific to a particular environment. Below is an example of developing a custom shimlet:

#### Built-in Example: Kubernetes Shimlet
ModelServeShim natively includes the Kubernetes Shimlet for deploying model services in Kubernetes environments. It implements the standard Shim interface and can convert abstract deployment requests into Kubernetes resource operations (such as creating Deployments and Services).

#### Step 1: Implement the Shim Interface
```go
package myshimlet

import (
    "context"
    "modserv-shim/internal/core/deploy"
)

// MyShimlet implements a custom environment adaptation plugin
type MyShimlet struct{}

// Create creates resources
func (s *MyShimlet) Create(ctx *deploy.Context) (string, error) {
    // Implement resource creation logic
    // Return resource ID
    return "resource-id", nil
}

// Status queries resource status
func (s *MyShimlet) Status(resourceID string) (deploy.Status, error) {
    // Implement resource status query logic
    return deploy.StatusRunning, nil
}

// Delete deletes resources
func (s *MyShimlet) Delete(resourceID string) error {
    // Implement resource deletion logic
    return nil
}

// GetResourceInfo gets detailed resource information
func (s *MyShimlet) GetResourceInfo(resourceID string) (map[string]interface{}, error) {
    // Implement detailed resource information retrieval logic
    return map[string]interface{}{"id": resourceID}, nil
}
```

#### Step 2: Register the Plugin
```go
package myshimlet

import (
    "modserv-shim/internal/core/plugin"
)

// init function is automatically called when the plugin is loaded
func init() {
    // Register the custom shimlet
    plugin.RegisterShimlet("my-shimlet", &MyShimlet{})
}
```

### Pipeline Development (Deployment Workflow Plugin)
Pipeline defines the specific steps and execution logic for model deployment. ModelServeShim implements Pipeline using the Builder pattern. Below is an example of developing a custom pipeline:

#### Built-in Example: OpenSourceLLM Pipeline
ModelServeShim natively includes the OpenSourceLLM Pipeline for open source large model deployment workflows. It is implemented using the Builder pattern and includes key steps such as generating service IDs, mapping model names to paths, applying service configurations, and exposing service endpoints, enabling users to quickly deploy open source large model services.

#### Step 1: Define Pipeline Step Functions
```go
package mypipeline

import (
    "modserv-shim/internal/core/pipeline"
    "modserv-shim/pkg/log"
)

// Define pipeline step functions of type func(*pipeline.Context) error

// validateModel validates model effectiveness
func validateModel(ctx *pipeline.Context) error {
    log.Info("Starting to validate model: %s", ctx.DeploySpec.ModelName)
    // Implement model validation logic
    return nil
}

// processConfig processes deployment configuration
func processConfig(ctx *pipeline.Context) error {
    log.Info("Processing deployment configuration")
    // Implement configuration processing logic
    return nil
}

// prepareResources prepares deployment resources
func prepareResources(ctx *pipeline.Context) error {
    log.Info("Preparing deployment resources")
    // Implement resource preparation logic
    return nil
}
```

#### Step 2: Create and Register Pipeline
```go
package mypipeline

import (
    "modserv-shim/internal/core/pipeline"
)

// init function is automatically called when the plugin is loaded
func init() {
    // Create and register custom pipeline using Builder pattern
    myCustomPipeline()
}

// myCustomPipeline creates a custom pipeline instance
func myCustomPipeline() *pipeline.Pipeline {
    // Use New() to create builder, Step() to add steps, BuildAndRegister() to complete construction and registration
    return pipeline.New("my-pipeline").
        Step(validateModel).
        Step(processConfig).
        Step(prepareResources).
        BuildAndRegister()
}
```

### Extended Example: Docker Shimlet
In addition to the built-in Kubernetes Shimlet, developers can implement Docker environment adaptation plugins to deploy model services in Docker containers. Docker Shimlet creates and manages containers through the Docker API, supporting complete lifecycle management of model services.

### Extended Example: Business Scenario Pipeline
Developers can create dedicated Pipelines based on specific business requirements. For example:
- **Multimodal Model Service Pipeline**: Add special validation steps for text and image processing, optimize GPU allocation strategies, configure dedicated inference parameters
- **Edge Deployment Pipeline**: Add resource limit checks, model quantization optimization, offline inference support and other special steps
- **Enterprise Security Pipeline**: Integrate identity verification, encrypted transmission, access control and other security enhancement features

### Plugin Integration Method

ModelServeShim implements plugin integration using Go's initialization registration mechanism, not through shared library compilation and hot loading.

#### Built-in Plugin Integration
Built-in plugins (such as Kubernetes Shimlet) are automatically registered into the framework through the `init()` function:
```go
// Example of K8sShimlet registration method
func init() {
    shimlet.Registry.AutoRegister(&K8sShimlet{})
}
```

#### Custom Plugin Integration
Custom plugins can be integrated into ModelServeShim through the following methods:

1. **Implement Standard Interfaces**: Implement the `Shimlet` or `Pipeline` interfaces as shown in the documentation
2. **Automatic Registration**: Complete automatic registration using the registry in the `init()` function
3. **Recompile**: Place the custom plugin code in the correct package path, then recompile the entire application

#### Plugin Selection and Configuration
Specify the plugins to use through command-line parameters or configuration files:
```bash
# Specify plugins via command line
./model-serve-shim --shimlet=k8s --pipeline=opensourcellm

# Specify plugins via configuration file
# Set in config.yaml
defaultShimlet: k8s
defaultPipeline: opensourcellm
```

## Configuration Instructions
ModelServeShim supports configuration through command-line parameters and configuration files:

### Command-line Parameters
```bash
./model-serve-shim --help

Usage of model-serve-shim:
  --port int              Service listening port (default: 8080)
  --config string         Configuration file path
  --shimlet string        Default loaded shimlet plugin
  --pipeline string       Default loaded pipeline plugin
  --plugin-dir string     Plugin directory path
  --log-level string      Log level (debug, info, warn, error) (default: "info")
```

### Configuration File
The configuration file uses YAML format:
```yaml
# config.yaml
service:
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s

plugins:
  defaultShimlet: k8s
  defaultPipeline: opensourcellm
  pluginDir: ./plugins
  preload: 
    - type: shimlet
      path: ./plugins/myshimlet.so
    - type: pipeline
      path: ./plugins/mypipeline.so

logging:
  level: info
  format: text
  output: stdout
```

## Contribution Guide
We welcome community contributions. Please read the following guidelines before contributing:

1. Fork the repository and create your own branch
2. Follow the project's code standards (use pre-commit for code style checks)
3. Ensure all tests pass before submitting code
4. Submit a Pull Request describing the changes made and the problems solved

## License
ModelServeShim is licensed under the Apache License 2.0.

## Contact Us
For questions or suggestions, please contact us through the following channels: