# ModelServeShim 

## 一句话介绍
极致轻量AI服务管控中间件，基于双插件化架构（`shimlet` 基础环境适配 + `pipeline` 部署流程扩展），原生内置集成`k8s-shimlet`(k8s环境支持) 与 `opensourcellm-pipeline`(开源模型部署流程))，支持跨环境部署与自定义流程，简化全生命周期管控。

## 命名释义
ModelServeShim（简称 modserv-shim）由两部分构成：

ModelServe：体现核心能力 —— 聚焦大模型（Model）的服务化部署与运维（Serve）；
Shim：源自技术术语 “适配层”，指通过抽象接口屏蔽底层环境差异，是实现跨环境统一管控的核心设计。

## 核心特性
| 特性分类                | 核心能力说明                                                                 | 关键价值                                                                 |
|-------------------------|------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| 插件化跨环境适配架构    | 1. 基于 `shim` 抽象层定义标准化接口，通过 `shimlet` 插件实现环境适配<br>2. 已实现 `k8sshimlet`（深度对接 K8s Deployment/StatefulSet 资源）<br>3. 新增环境仅需开发专属 `shimlet`，核心逻辑与配置完全复用 | 一套管控逻辑覆盖多部署场景，环境扩展无需修改核心代码，开发成本显著降低      |
| 可定制化部署流水线      | 1. 基于 `pipeline` 插件化设计，默认提供 `opensourcellm pipeline`（开源大模型部署流程）<br>2. 支持自定义 `pipeline` 替换默认流程，可新增量化、权限校验等个性化步骤<br>3. 流水线步骤支持配置重试策略与超时控制 | 兼顾开源模型快速部署与业务场景定制化需求，流程扩展灵活无侵入                |
| 轻量架构与便捷部署      | 1. 单二进制文件交付（≈15MB），无外部依赖，解压即可启动<br>2. 支持插件热加载，新增 `shimlet`/`pipeline` 无需重启服务<br>3. 配置简洁，仅需指定环境插件与流程插件即可运行 | 部署门槛低，运维成本小，适配中小团队快速落地需求                          |
| 全链路状态可视与监控    | 1. FSM 状态机驱动，实时展示「初始化→流水线执行→环境部署→健康检查→运行中」全流转<br>2. 每个状态与步骤关联详细日志（操作人、耗时、参数）<br>3. 暴露 Prometheus 指标：部署成功率、资源使用率、流水线耗时等 | 状态可追溯，故障可快速定位，服务可用性全程可控                            |


## 技术架构
采用“核心逻辑+双插件”解耦设计，兼顾稳定性与扩展性：

![img.png](../../img.png)

## 快速上手（K8s + 开源LLM默认流水线）
### 1. 环境准备
- K8s 集群（v1.20+），已配置 GPU 调度（nvidia-device-plugin）
- 本地 `kubectl` 可访问集群（配置 KUBECONFIG）

### 2. 下载安装
```bash
# 下载单文件二进制（Linux x86_64）
wget https://github.com/your-org/ModelServeShim/releases/latest/download/model-serve-shim
chmod +x model-serve-shim

# 验证安装（无依赖检查）
./model-serve-shim --version
# 输出示例：ModelServeShim v1.0.0 (commit: abc123)
```

### 3. 启动服务（加载默认插件）
```bash
# 启用 K8s shimlet 与开源LLM默认 pipeline
./model-serve-shim --port=8080 \
  --shimlet=k8s \
  --pipeline=opensourcellm
```

### 4. 部署开源大模型（Llama-3-8B 示例）
```bash
curl -X POST http://localhost:8080/api/v1/modserv/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "modelName": "llama-3-8b",
    "modelFile": "/models/llama-3-8b",
    "resourceRequirements": {
      "acceleratorType": "NVIDIA H20",
      "acceleratorCount": 1,
      "cpu": "4",
      "memory": "16Gi"
    },
    "replicaCount": 1
  }'

# 响应示例（获取 serviceId 跟踪进度）
{
  "code": 0,
  "data": {
    "serviceId": "llama-3-8b-123e4567-e89b-12d3-a456-426614174000",
    "currentStep": "model-validation",
    "totalSteps": ["model-validation", "config-rendering", "resource-deployment", "health-check"]
  }
}
```

### 5. 查看部署状态与流水线进度
```bash
# 查看整体状态
curl http://localhost:8080/api/v1/modserv/llama-3-8b-123e4567-e89b-12d3-a456-426614174000

# 查看流水线详细执行日志
curl http://localhost:8080/api/v1/modserv/llama-3-8b-123e4567-e89b-12d3-a456-426614174000/pipeline/logs
```


## 插件扩展示例
### 1. 新增 shimlet（适配 Docker 环境）
只需实现 `Shim` 抽象接口，即可接入新环境：
``` go
// dockershimlet 核心实现
type DockerShimlet struct{}

// 实现创建资源接口
func (d *DockerShimlet) Create(ctx *deploy.Context) (string, error) {
    // 调用 Docker SDK 创建容器
    container, err := dockerClient.ContainerCreate(
        context.Background(),
        &container.Config{Image: ctx.ModelImage},
        &container.HostConfig{Resources: getDockerResources(ctx.Requirements)},
        nil, nil, ctx.ModelName)
    if err != nil {
        return "", fmt.Errorf("create docker container failed: %v", err)
    }
    // 启动容器
    return container.ID, dockerClient.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
}

// 实现查询状态接口
func (d *DockerShimlet) Status(resourceID string) (deploy.Status, error) {
    // 查询 Docker 容器状态
    inspect, err := dockerClient.ContainerInspect(context.Background(), resourceID)
    if err != nil {
        return deploy.StatusFailed, err
    }
    if inspect.State.Running {
        return deploy.StatusRunning, nil
    }
    return deploy.StatusStopped, nil
}
```

### 2. 自定义 pipeline（增加模型量化步骤）
实现 `Pipeline` 接口，替换默认部署流程：
``` go
// 带量化步骤的自定义流水线
type QuantLLMpipeline struct{}

// 定义流水线步骤
func (p *QuantLLMpipeline) Steps() []string {
    return []string{"model-validation", "model-quantization", "config-rendering", "resource-deployment", "health-check"}
}

// 实现步骤执行逻辑
func (p *QuantLLMpipeline) RunStep(step string, ctx *pipeline.Context) error {
    switch step {
    case "model-quantization":
        // 调用量化工具（如 GGUF）处理模型
        return quantizeModel(ctx.ModelFile, ctx.QuantConfig)
    case "model-validation":
        return validateModel(ctx.ModelFile)
    // 实现其他步骤...
    default:
        return fmt.Errorf("unsupported step: %s", step)
    }
}
```


## 常用 API 参考
| 操作类型         | HTTP 方法 | 接口路径                          | 说明                                  |
|------------------|-----------|-----------------------------------|---------------------------------------|
| 部署服务         | POST      | `/api/v1/modserv/deploy`          | 使用指定插件创建模型服务实例          |
| 查询服务状态     | GET       | `/api/v1/modserv/{serviceId}`     | 查看服务整体状态与资源信息            |
| 查询流水线进度   | GET       | `/api/v1/modserv/{id}/pipeline`   | 查看流水线步骤执行情况                |
| 列出可用插件     | GET       | `/api/v1/plugins`                 | 查看已加载的shimlet与pipeline      |
| 加载新插件       | POST      | `/api/v1/plugins/load`            | 热加载自定义插件（无需重启服务）       |


## 插件生态路线图
| 插件类型       | 现有实现                | 规划扩展                  |
|----------------|-------------------------|---------------------------|
| shimlet        | k8sshimlet              | dockershimlet、edgelet    |
| pipeline    | opensourcellm pipeline | privatellm pipeline、quant-pipeline |

## 🛠️ 代码规范

本项目使用 [pre-commit](https://pre-commit.com) 自动检查代码风格，确保提交的代码格式统一。

[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://pre-commit.com/)
## 许可证
Apache License 2.0