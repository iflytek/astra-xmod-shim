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
	"path/filepath"
	"strings"
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
// 主要修复：
// 1. 移除 model 路径的 'local:' 前缀（vLLM 不支持）
// 2. 确保 modelDirPath 是目录（非文件）
// 3. 正确设置 hostPath volume type
// 4. 确保 volumeMount 路径与 --model 参数一致
func (k *K8sShimlet) Apply(deploySpec *dto.DeploySpec) (string, error) {
	// 1. 创建容器配置
	deploymentName := utils.ModelNameToDeploymentName(deploySpec.ModelName) + "-" + deploySpec.ServiceId
	mainContainerName := utils.ModelNameToDeploymentName(deploySpec.ModelName)
	imageName := "artifacts.iflytek.com/docker-private/aiaas/vllm-openai:v0.4.2"
	// 使用映射后的模型路径（通过pipeline的mapModelNameToPath步骤设置）
	modelDirPath := deploySpec.ModelFileDir

	// 如果ModelFileDir为空，直接报错
	if modelDirPath == "" {
		return "", errors.New("模型路径不能为空，请提供有效的模型名称")
	}

	// ✅ 关键修复：确保 modelDirPath 是目录，不是文件
	// 如果传入的是模型文件（如 .bin, .safetensors），取其父目录
	if strings.HasSuffix(strings.ToLower(modelDirPath), ".bin") ||
		strings.HasSuffix(strings.ToLower(modelDirPath), ".safetensors") ||
		strings.HasSuffix(strings.ToLower(modelDirPath), ".pt") ||
		strings.HasSuffix(strings.ToLower(modelDirPath), ".gguf") {
		modelDirPath = filepath.Dir(modelDirPath)
	}

	// 再次校验路径是否有效
	if modelDirPath == "" || modelDirPath == "." || modelDirPath == "/" {
		return "", errors.New("解析后的模型路径无效")
	}

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
	// ✅ 移除 local: 前缀，vLLM 需要纯路径
	envVars := []*corev1apply.EnvVarApplyConfiguration{
		{
			Name:  &[]string{"MODEL"}[0],
			Value: &modelDirPath, // ✅ 直接使用路径，不要加 "local:"
		},
		{
			Name:  &[]string{"SERVING_ENGINE"}[0],
			Value: &[]string{"openai"}[0],
		},
		{
			Name:  &[]string{"PORT"}[0],
			Value: &portStr,
		},
		// ✅ 强制离线模式（防止 HF 联网下载）
		{
			Name:  &[]string{"TRANSFORMERS_OFFLINE"}[0],
			Value: &[]string{"1"}[0],
		},
		// ✅ 避免 HF 写默认目录失败
		{
			Name:  &[]string{"HF_HOME"}[0],
			Value: &[]string{"/tmp"}[0],
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
	container.WithPorts(
		corev1apply.ContainerPort().
			WithName("http").
			WithContainerPort(randomPort),
	)

	// ✅ 关键修复：添加 args，确保 vLLM 监听指定端口并使用本地模型路径
	// ✅ 在 --model 参数中移除 'local:' 前缀
	container.WithArgs(
		"--host=0.0.0.0",
		"--port="+portStr,
		"--model="+modelDirPath, // ✅ 纯路径
		"--dtype=auto",
		"--trust-remote-code", // ✅ Qwen 必须加
	)

	// 6. 构建完整的Deployment配置
	deploymentApply := &appsv1apply.DeploymentApplyConfiguration{}
	deploymentApply.WithAPIVersion("apps/v1")
	deploymentApply.WithKind("Deployment")
	deploymentApply.WithName(deploymentName)
	deploymentApply.WithNamespace("default")
	deploymentApply.WithLabels(map[string]string{
		"app":        deploySpec.ServiceId,
		"managed-by": "modserv-shim",
	})

	// 设置Spec
	spec := &appsv1apply.DeploymentSpecApplyConfiguration{}
	spec.WithReplicas(int32(deploySpec.ReplicaCount))

	// 设置选择器
	selector := &metav1apply.LabelSelectorApplyConfiguration{}
	selector.WithMatchLabels(map[string]string{"app": deploySpec.ServiceId})
	spec.WithSelector(selector)

	// 设置Pod模板
	template := &corev1apply.PodTemplateSpecApplyConfiguration{}
	template.WithLabels(map[string]string{"app": deploySpec.ServiceId})

	// 设置Pod Spec
	podSpec := &corev1apply.PodSpecApplyConfiguration{}
	podSpec.WithHostNetwork(true) // 启用hostNetwork模式

	// ✅ 添加容忍所有污点
	podSpec.WithTolerations(
		corev1apply.Toleration().
			WithKey("").
			WithOperator(corev1.TolerationOpExists),
	)

	// ✅ 正确方式：使用链式调用添加 Volume（hostPath 挂载模型目录）
	podSpec.WithVolumes(
		corev1apply.Volume().
			WithName("models").
			WithHostPath(
				corev1apply.HostPathVolumeSource().
					WithPath(modelDirPath). // 宿主机路径
					WithType(corev1.HostPathDirectory), // ✅ 明确指定为目录
			),
	)

	// ✅ 添加 VolumeMount：将宿主机目录挂载到容器内
	// ✅ mountPath 必须与 --model 参数和 MODEL 环境变量的值完全一致
	container.WithVolumeMounts(
		corev1apply.VolumeMount().
			WithName("models").
			WithMountPath(modelDirPath), // 容器内路径
	)

	// ✅ 将容器加入 PodSpec
	podSpec.WithContainers(container)

	// ✅ 设置 Pod 模板 Spec
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

// ptr 是一个辅助函数，用于创建 *string
func ptr(s string) *string                                                { return &s }
func (k *K8sShimlet) Delete(resourceId string) error                      { return nil }
func (k *K8sShimlet) Status(resourceId string) (*dto.DeployStatus, error) { return nil, nil }
func (k *K8sShimlet) Description() string                                 { return "k8s shimlet" }
