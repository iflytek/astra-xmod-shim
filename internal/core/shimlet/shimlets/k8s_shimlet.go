package shimlets

import (
	"context"
	"errors"
	"fmt"
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/shimlet"
	cfg "modserv-shim/internal/dto/config"
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/pkg/k8s"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// ApplyConfigurations 包
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
)

// 编译时检查 确保实现 shimlet 接口
var _ shimlet.Shimlet = (*K8sShimlet)(nil)

func init() {
	shimlet.Registry.AutoRegister(&K8sShimlet{})
}

type K8sShimlet struct {
	client *k8s.K8sClient
}

func (k *K8sShimlet) ID() string { return "k8s" }

func (k *K8sShimlet) InitWithConfig(confPath string) error {
	k8sCfg, err := config.GetConfFromFileDir[cfg.K8sConfig](confPath)
	if err != nil {
		return err
	}
	client, err := k8s.NewK8sClient(k8sCfg)
	if err != nil {
		return errors.New("初始化K8s客户端失败")
	}
	k.client = client
	return nil
}

// Apply 使用 Server-Side Apply 方法部署应用（修复版）
func (k *K8sShimlet) Apply(deploySpec *dto.DeploySpec) (string, error) {
	// 1. 创建容器配置
	container := &corev1apply.ContainerApplyConfiguration{}
	container.WithName(deploySpec.ServiceId)
	// 使用 vllm 稳定版本，支持 OpenAI-like API
	container.WithImage("vllm/vllm-openai:v0.4.2")
	container.WithImagePullPolicy(corev1.PullIfNotPresent)

	// 2. 添加资源需求
	if deploySpec.ResourceRequirements != nil {
		resources := &corev1apply.ResourceRequirementsApplyConfiguration{}
		// 不设置任何CPU和内存限制
		resources.WithRequests(corev1.ResourceList{})
		resources.WithLimits(corev1.ResourceList{})

		// 添加GPU加速卡配置（如果有）
		if deploySpec.ResourceRequirements.AcceleratorType != "" && deploySpec.ResourceRequirements.AcceleratorCount > 0 {
			// 创建只包含GPU的limits对象
			limits := corev1.ResourceList{}

			// 添加GPU资源
			limits[corev1.ResourceName(deploySpec.ResourceRequirements.AcceleratorType)] = 
				resource.MustParse(fmt.Sprintf("%d", deploySpec.ResourceRequirements.AcceleratorCount))
			resources.WithLimits(limits)
		}

		container.WithResources(resources)
	}

	// 3. 添加环境变量 - 为vllm配置OpenAI兼容模式
	// 基础环境变量
	envVars := []*corev1apply.EnvVarApplyConfiguration{
		{
			Name:  &[]string{"MODEL"}[0],
			Value: &[]string{"facebook/opt-125m"}[0], // 默认模型，可根据需要更改
		},
		{
			Name:  &[]string{"SERVING_ENGINE"}[0],
			Value: &[]string{"openai"}[0], // 启用OpenAI兼容模式
		},
		{
			Name:  &[]string{"PORT"}[0],
			Value: &[]string{"8000"}[0], // API服务端口
		},
	}

	// 添加用户自定义环境变量
	for _, env := range deploySpec.Env {
		envVar := &corev1apply.EnvVarApplyConfiguration{}
		envVar.WithName(env.Key)
		envVar.WithValue(env.Val)
		envVars = append(envVars, envVar)
	}

	container.WithEnv(envVars...)

	// 4. 添加端口配置
	container.WithPorts(&corev1apply.ContainerPortApplyConfiguration{Name: &[]string{"http"}[0], ContainerPort: &[]int32{8000}[0]})

	// 5. 构建完整的Deployment配置
	deploymentApply := &appsv1apply.DeploymentApplyConfiguration{}
	deploymentApply.WithName(deploySpec.ServiceId)
	deploymentApply.WithNamespace("default")
	deploymentApply.WithLabels(map[string]string{
		"app":        deploySpec.ServiceId,
		"managed-by": "modserv-shim",
	})

	// 6. 设置Spec
	spec := &appsv1apply.DeploymentSpecApplyConfiguration{}
	spec.WithReplicas(int32(deploySpec.ReplicaCount))

	// 7. 设置选择器
	selector := &metav1apply.LabelSelectorApplyConfiguration{}
	selector.WithMatchLabels(map[string]string{"app": deploySpec.ServiceId})
	spec.WithSelector(selector)

	// 8. 设置Pod模板
	template := &corev1apply.PodTemplateSpecApplyConfiguration{}
	template.WithLabels(map[string]string{"app": deploySpec.ServiceId})

	// 9. 设置Pod Spec
	podSpec := &corev1apply.PodSpecApplyConfiguration{}
	podSpec.WithContainers(container)
	template.WithSpec(podSpec)

	// 10. 将模板添加到spec
	spec.WithTemplate(template)
	deploymentApply.WithSpec(spec)

	// 11. 执行 Server-Side Apply
	result, err := k.client.GetClientSet().AppsV1().Deployments("default").Apply(
		context.Background(),
		deploymentApply,
		metav1.ApplyOptions{FieldManager: "modserv-shim", Force: true},
	)
	if err != nil {
		return "", fmt.Errorf("部署应用失败: %w", err)
	}

	return fmt.Sprintf("应用 %s/%s 部署成功", result.Namespace, result.Name), nil
}

// 删除、状态查询等方法保持不变
func (k *K8sShimlet) Delete(resourceId string) error                      { return nil }
func (k *K8sShimlet) Status(resourceId string) (*dto.DeployStatus, error) { return nil, nil }
func (k *K8sShimlet) Description() string                                 { return "k8s shimlet" }