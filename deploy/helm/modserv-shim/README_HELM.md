# modserv-shim Helm Chart 使用说明

## 概述

本Helm Chart用于部署modserv-shim应用到Kubernetes集群，使用主机网络模式，并挂载必要的配置文件和模型目录。

## 主要特性

- 使用主机网络模式运行
- 挂载项目配置文件
- 挂载主机模型目录`/mnt/maasmodels/`
- 支持k8sshimlet配置文件挂载

## 配置修改说明

### 关键配置项

1. **主机网络模式**
   - 已启用`hostNetwork: true`
   - 配置了`dnsPolicy: ClusterFirstWithHostNet`以确保DNS解析正常

2. **卷挂载配置**
   - 项目配置目录: 挂载到容器的`/app/conf`
   - 模型目录: 挂载主机的`/mnt/maasmodels/`到容器相同路径

3. **服务配置**
   - 当使用主机网络模式时，设置了`clusterIP: None`

## 部署步骤

### 前提条件

- 已安装Helm 3.x
- 已配置kubectl连接到目标Kubernetes集群
- 主机上已存在配置目录和模型目录

### 部署命令

```bash
# 进入Helm chart目录
cd /Users/haoxuanli/Documents/GitHub/iflytek/modserv-shim/deploy/helm

# 安装或升级应用
helm upgrade --install modserv-shim modserv-shim/ -f modserv-shim/values.yaml

# 验证部署
kubectl get pods -l app.kubernetes.io/name=modserv-shim
```

## 自定义配置

如需修改默认配置，可以通过以下方式：

1. **修改values.yaml文件**
   ```bash
   vi modserv-shim/values.yaml
   ```

2. **使用自定义values文件**
   ```bash
   helm upgrade --install modserv-shim modserv-shim/ -f your-custom-values.yaml
   ```

## 重要说明

- 主机网络模式下，容器将直接使用主机的网络命名空间，需要确保主机端口不被占用
- 配置文件和模型目录的路径需根据实际环境调整
- 本Chart假设应用程序的配置文件在容器内的`/app/conf`目录下读取

## 故障排查

- **端口冲突**：如果部署失败，检查主机端口是否被其他应用占用
- **权限问题**：确保容器有足够的权限访问挂载的目录
- **配置错误**：验证配置文件路径和内容是否正确

## 卸载

```bash
helm uninstall modserv-shim
```