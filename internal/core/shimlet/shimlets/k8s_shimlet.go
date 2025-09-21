package shimlets

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/shimlet"
	cfg "modserv-shim/internal/dto/config"
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/pkg/k8s"
	"modserv-shim/pkg/utils"

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
	if k == nil {
		return nil
	}
	k.client = client
	return nil
}

// Apply 使用 Server-Side Apply 方法部署应用（修复版）
// Apply 使用 Server-Side Apply 方法部署应用（修复版）
func (k *K8sShimlet) Apply(deploySpec *dto.DeploySpec) (string, error) {
	// 1. 创建容器配置

	deploymentName := utils.ModelNameToDeploymentName(deploySpec.ModelName) + "-" + deploySpec.ServiceId
	mainContainerName := utils.ModelNameToDeploymentName(deploySpec.ModelName)
	imageName := "artifacts.iflytek.com/docker-private/aiaas/vllm-openai:v0.4.2"

	container := &corev1apply.ContainerApplyConfiguration{}
	container.WithName(mainContainerName)
	container.WithImage(imageName)
	container.WithImagePullPolicy(corev1.PullIfNotPresent)

	// 2. 添加资源需求
	if deploySpec.ResourceRequirements != nil {
		resources := &corev1apply.ResourceRequirementsApplyConfiguration{}
		resources.WithRequests(corev1.ResourceList{})
		resources.WithLimits(corev1.ResourceList{})

		if deploySpec.ResourceRequirements.AcceleratorType != "" && deploySpec.ResourceRequirements.AcceleratorCount > 0 {
			limits := corev1.ResourceList{}
			limits[corev1.ResourceName(deploySpec.ResourceRequirements.AcceleratorType)] =
				resource.MustParse(fmt.Sprintf("%d", deploySpec.ResourceRequirements.AcceleratorCount))
			resources.WithLimits(limits)
		}

		container.WithResources(resources)
	}

	// 3. 生成随机端口（使用NodePort范围30000-32767）
	randomPort := rand.Int31n(2768) + 30000 // 随机端口范围: 30000-32767
	portStr := fmt.Sprintf("%d", randomPort)

	// 4. 添加环境变量
	envVars := []*corev1apply.EnvVarApplyConfiguration{
		{
			Name:  &[]string{"MODEL"}[0],
			Value: &[]string{"facebook/opt-125m"}[0],
		},
		{
			Name:  &[]string{"SERVING_ENGINE"}[0],
			Value: &[]string{"openai"}[0],
		},
		{
			Name:  &[]string{"PORT"}[0],
			Value: &[]string{portStr}[0],
		},
	}

	for _, env := range deploySpec.Env {
		envVar := &corev1apply.EnvVarApplyConfiguration{}
		envVar.WithName(env.Key)
		envVar.WithValue(env.Val)
		envVars = append(envVars, envVar)
	}

	container.WithEnv(envVars...)

	// 5. 添加端口配置，使用随机端口
	container.WithPorts(&corev1apply.ContainerPortApplyConfiguration{Name: &[]string{"http"}[0], ContainerPort: &[]int32{randomPort}[0]})

	// 5. 构建完整的Deployment配置（补充 apiVersion 和 kind）
	deploymentApply := &appsv1apply.DeploymentApplyConfiguration{}
	// 关键修复：添加 API版本和资源类型（必填！）
	deploymentApply.WithAPIVersion("apps/v1") // Deployment 的标准 API 版本
	deploymentApply.WithKind("Deployment")    // 资源类型为 Deployment
	// 原有字段不变
	deploymentApply.WithName(deploymentName)
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

	// 9. 设置Pod Spec，启用hostNetwork
	podSpec := &corev1apply.PodSpecApplyConfiguration{}
	podSpec.WithHostNetwork(true) // 启用hostNetwork模式
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

	return fmt.Sprintf("应用 %s/%s 部署成功，使用hostNetwork并暴露端口 %d", result.Namespace, result.Name, randomPort), nil
}

func (k *K8sShimlet) Delete(resourceId string) error                      { return nil }
func (k *K8sShimlet) Status(resourceId string) (*dto.DeployStatus, error) { return nil, nil }
func (k *K8sShimlet) Description() string                                 { return "k8s shimlet" }