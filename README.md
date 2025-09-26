<div align="center">
<img src="xmod-shim.svg" alt="Astra-mod-shim Logo" width="600" />
<br>

[![License](https://img.shields.io/github/license/iflytek/astra-xmod-shim)](LICENSE)
[![Release](https://img.shields.io/github/v/release/iflytek/astra-xmod-shim?include_prereleases)](https://github.com/iflytek/astra-xmod-shim/releases)
[![CI Status](https://img.shields.io/github/actions/workflow/status/iflytek/astra-xmod-shim/ci.yml?branch=main)](https://github.com/iflytek/astra-xmod-shim/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/iflytek/astra-xmod-shim)](go.mod)
[![Coverage](https://img.shields.io/codecov/c/github/iflytek/astra-xmod-shim)](https://codecov.io/gh/iflytek/astra-xmod-shim)
![Multi-Arch](https://img.shields.io/badge/Multi--Arch-linux%2Famd64%20%7C%20linux%2Farm64-blue?logo=docker)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Native-blue?logo=kubernetes&logoColor=white)](docs/k8s.md)
[![Helm](https://img.shields.io/badge/Helm-Chart-blue?logo=helm&logoColor=white)](charts/)
[![Cloud Native](https://img.shields.io/badge/Cloud-Native-blue?logo=cloudnative&logoColor=white)](https://cncf.io)
[![Metrics](https://img.shields.io/badge/Metrics-Prometheus-green?logo=prometheus)](docs/metrics.md)
[![Contributors](https://img.shields.io/github/contributors/iflytek/astra-xmod-shim)](https://github.com/iflytek/astra-xmod-shim/graphs/contributors)
[![Stars](https://img.shields.io/github/stars/iflytek/astra-xmod-shim?style=social)](https://github.com/iflytek/astra-xmod-shim)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

</div>



Astra-xmod-shim 是一款轻量级 AI 模型服务管控中间件，采用插件化架构设计，旨在简化大模型服务的部署、运维与全生命周期管理。

通过 **面向抽象的运行时适配（Shimlet）** 与 **函数式可组合的部署流程（Pipeline）**
，系统将“在哪运行”与“如何部署”彻底解耦，实现跨环境统一管控，支持快速集成新平台与定制化流程，真正达成“**一次定义，随处部署**
”的敏捷能力。

<span style="font-size:0.9em; color:#586375;">**Language**: [English](README_en.md) | **简体中文**</span>

## 项目概述

Astra-xmod-shim 是一款轻量级 AI 模型服务管控中间件，采用插件化架构设计，旨在简化大模型服务的部署、运维与全生命周期管理。

通过 面向抽象的运行时适配（Shimlet） 与
函数式可组合的部署流程（Pipeline），系统将“在哪运行”与“如何部署”彻底解耦，实现跨环境统一管控，支持快速集成新平台与定制化流程，真正达成“一次定义，随处部署”的敏捷能力。

## 🌟 核心特性

- **插件化环境抽象（Shimlet）**  
  基于接口抽象实现运行时解耦，支持 K8s、Docker 等环境通过插件无缝切换

- **函数式部署编排（Pipeline）**  
  采用函数链式编排，灵活定义部署流程，支持验证、配置、启动等阶段可扩展

- **轻量单体架构**  
  单二进制交付，无外部依赖，适用于边缘、本地及云原生部署场景

- **状态机驱动的全生命周期管理**  
  基于有限状态机（FSM）精确控制服务状态流转，支持部署、运行、更新、销毁的可靠追踪

- **事件驱动的可观测架构**  
  通过 EventBus 解耦核心与监控、日志、追踪组件，支持异步状态同步与扩展

## 🏗️ 技术架构

ModelServeShim 采用“核心引擎 + 双插件”的解耦架构，通过抽象层与流程引擎分离关注点，实现高可扩展性与环境无关性。

![架构示意图](img.png)

- **核心引擎（Core Engine）**  
  系统中枢，负责服务生命周期调度、API 接管与状态协调。基于有限状态机（FSM）管理模型服务的状态流转，确保操作的确定性与可观测性。

- **Shimlet（运行时适配插件）**  
  实现 `shim.Runtime` 接口，封装底层环境（如 Kubernetes、Docker）的资源操作。通过接口抽象实现运行时解耦，支持多环境无缝切换。

- **Pipeline（部署流程插件）**  
  由一系列函数式步骤（`pipeline.Step`）组成，定义模型部署的执行流程。支持阶段化编排（如校验、配置生成、资源创建），可灵活扩展。

- **事件总线（EventBus）**  
  耦合核心与观测组件，异步广播服务状态变更事件，支撑日志、监控、审计等外接系统。

## 快速开始

### 环境要求

- Go 1.20+（开发环境）
- 目标环境（如 K8s v1.20+，如需使用 K8s shimlet）

### 安装

```bash
# 下载二进制文件（Linux x86_64）
wget https://github.com/iflytek/modserv-shim/releases/latest/download/model-serve-shim
chmod +x model-serve-shim

# 或从源码构建
git clone https://github.com/iflytek/modserv-shim.git
cd modserv-shim
make build
```

### 基本使用

```bash
# 启动服务，加载 K8s shimlet 和开源 LLM 部署流程
./model-serve-shim --port=8080 \    
  --shimlet=k8s \                  
  --pipeline=opensourcellm          
```

## API 参考

### 部署模型服务

```bash
curl -X POST http://localhost:8080/api/v1/modserv/deploy \   
  -H "Content-Type: application/json" \                      
  -d '{                                                      
    "modelName": "example-model",                         
    "modelFile": "/path/to/model",                        
    "resourceRequirements": {                              
      "acceleratorType": "NVIDIA GPU",                    
      "acceleratorCount": 1,                               
      "cpu": "4",                                         
      "memory": "16Gi"                                    
    },                                                       
    "replicaCount": 1                                       
  }'                                                         
```

### 查询服务状态

```bash
curl http://localhost:8080/api/v1/modserv/{serviceId}
```

### 列出已加载插件

```bash
curl http://localhost:8080/api/v1/plugins
```

## 插件开发指南

### Shimlet 开发（环境适配插件）

Shimlet 负责将抽象的部署请求转换为具体环境的操作。以下是开发自定义 shimlet 的示例：

#### 内置示例：Kubernetes Shimlet

ModelServeShim 原生内置了 Kubernetes Shimlet，用于在 Kubernetes 环境中部署模型服务。它实现了标准的 Shim 接口，能够将抽象部署请求转换为
Kubernetes 的资源操作（如创建 Deployment 和 Service 等）。

#### 步骤 1：实现 Shim 接口

```go
package myshimlet

import (
	"context"
	"modserv-shim/internal/core/deploy"
)

// MyShimlet 实现自定义环境适配插件
type MyShimlet struct{}

// Create 创建资源
func (s *MyShimlet) Create(ctx *deploy.Context) (string, error) {
	// 实现创建资源的逻辑
	// 返回资源 ID
	return "resource-id", nil
}

// Status 查询资源状态
func (s *MyShimlet) Status(resourceID string) (deploy.Status, error) {
	// 实现查询资源状态的逻辑
	return deploy.StatusRunning, nil
}

// Delete 删除资源
func (s *MyShimlet) Delete(resourceID string) error {
	// 实现删除资源的逻辑
	return nil
}

// GetResourceInfo 获取资源详细信息
func (s *MyShimlet) GetResourceInfo(resourceID string) (map[string]interface{}, error) {
	// 实现获取资源详细信息的逻辑
	return map[string]interface{}{"id": resourceID}, nil
}
```

#### 步骤 2：注册插件

```go
package myshimlet

import (
	"modserv-shim/internal/core/plugin"
)

// init 函数在插件加载时自动调用
func init() {
	// 注册自定义 shimlet
	plugin.RegisterShimlet("my-shimlet", &MyShimlet{})
}
```

### Pipeline 开发（部署流程插件）

Pipeline 定义了模型部署的具体步骤和执行逻辑。ModelServeShim 使用 Builder 模式实现 Pipeline，以下是开发自定义 pipeline 的示例：

#### 内置示例：OpenSourceLLM Pipeline

ModelServeShim 原生内置了 OpenSourceLLM Pipeline，用于开源大模型的部署流程。它采用 Builder
模式实现，包含生成服务ID、映射模型名称到路径、应用服务配置和暴露服务端点等关键步骤，使用户能够快速部署开源大模型服务。

#### 步骤 1：定义 Pipeline 步骤函数

```go
package mypipeline

import (
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/pkg/log"
)

// 定义 pipeline 步骤函数，类型为 func(*pipeline.Context) error

// validateModel 验证模型有效性
func validateModel(ctx *pipeline.Context) error {
	log.Info("开始验证模型: %s", ctx.DeploySpec.ModelName)
	// 实现模型验证逻辑
	return nil
}

// processConfig 处理部署配置
func processConfig(ctx *pipeline.Context) error {
	log.Info("处理部署配置")
	// 实现配置处理逻辑
	return nil
}

// prepareResources 准备部署资源
func prepareResources(ctx *pipeline.Context) error {
	log.Info("准备部署资源")
	// 实现资源准备逻辑
	return nil
}
```

#### 步骤 2：创建并注册 Pipeline

```go
package mypipeline

import (
	"modserv-shim/internal/core/pipeline"
)

// init 函数在插件加载时自动调用
func init() {
	// 使用 Builder 模式创建并注册自定义 pipeline
	myCustomPipeline()
}

// myCustomPipeline 创建自定义 pipeline 实例
func myCustomPipeline() *pipeline.Pipeline {
	// 使用 New() 创建 builder，Step() 添加步骤，BuildAndRegister() 完成构建并注册
	return pipeline.New("my-pipeline").
		Step(validateModel).
		Step(processConfig).
		Step(prepareResources).
		BuildAndRegister()
}
```

### 扩展示例：Docker Shimlet

除了内置的Kubernetes Shimlet外，开发者还可以实现Docker环境适配插件，将模型服务部署到Docker容器中。Docker Shimlet通过Docker
API创建和管理容器，支持模型服务的完整生命周期管理。

### 扩展示例：业务场景 Pipeline

开发者可以根据具体业务需求创建专用的Pipeline。例如：

- **多模态模型服务Pipeline**：增加针对文本和图像处理的特殊验证步骤、优化GPU分配策略、配置专用推理参数
- **边缘部署Pipeline**：添加资源限制检查、模型量化优化、离线推理支持等特殊步骤
- **企业级安全Pipeline**：集成身份验证、加密传输、访问控制等安全增强功能

### 插件集成方式

ModelServeShim 使用 Go 语言的初始化注册机制实现插件集成，而不是通过共享库编译和热加载。

#### 内置插件集成

内置插件（如 Kubernetes Shimlet）通过在 `init()` 函数中自动注册到框架中：

```go
// K8sShimlet 的注册方式示例
func init() {
shimlet.Registry.AutoRegister(&K8sShimlet{})
}
```

#### 自定义插件集成

自定义插件可以通过以下方式集成到 ModelServeShim 中：

1. **实现标准接口**：按照文档中示例实现 `Shimlet` 或 `Pipeline` 接口
2. **自动注册**：在 `init()` 函数中使用注册表完成自动注册
3. **重新编译**：将自定义插件代码放在正确的包路径下，然后重新编译整个应用程序

#### 插件选择与配置

通过命令行参数或配置文件指定要使用的插件：

```bash
# 通过命令行指定插件
./model-serve-shim --shimlet=k8s --pipeline=opensourcellm

# 通过配置文件指定插件
# config.yaml 中设置
defaultShimlet: k8s
defaultPipeline: opensourcellm
```

## 配置说明

ModelServeShim 支持通过命令行参数和配置文件进行配置：

### 命令行参数

```bash
./model-serve-shim --help

Usage of model-serve-shim:
  --port int              服务监听端口 (默认: 8080)
  --config string         配置文件路径
  --shimlet string        默认加载的 shimlet 插件
  --pipeline string       默认加载的 pipeline 插件
  --plugin-dir string     插件目录路径
  --log-level string      日志级别 (debug, info, warn, error) (默认: "info")
```

### 配置文件

配置文件采用 YAML 格式：

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

## 贡献指南

我们欢迎社区贡献，贡献前请阅读以下指南：

1. Fork 仓库并创建自己的分支
2. 遵循项目代码规范（使用 pre-commit 进行代码风格检查）
3. 提交代码前确保通过所有测试
4. 提交 Pull Request，描述清楚所做的变更和解决的问题

## 🌟 Star 历史

<div align="center">
  <img src="https://api.star-history.com/svg?repos=iflytek/astra-xmod-shim
&type=Date" alt="Star History Chart" width="600">
</div>

## 许可证

ModelServeShim 使用 Apache License 2.0 许可证。

## 联系我们

如有问题或建议，请通过以下方式联系我们：

- GitHub Issues: https://github.com/iflytek/modserv-shim/issues
- Email: hxli28@iflytek.com