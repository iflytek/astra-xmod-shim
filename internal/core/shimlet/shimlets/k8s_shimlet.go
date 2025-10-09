package shimlets

import (
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/internal/core/shimlet"
	cfg "astron-xmod-shim/internal/dto/config"
	dto "astron-xmod-shim/internal/dto/deploy"
	"astron-xmod-shim/pkg/k8s"
	"astron-xmod-shim/pkg/log"
	"astron-xmod-shim/pkg/utils"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
)

// Ensure K8sShimlet implements the Shimlet interface at compile time
var _ shimlet.Shimlet = (*K8sShimlet)(nil)

func init() {
	shimlet.Registry.AutoRegister(&K8sShimlet{})
}

// K8sShimlet provides a Kubernetes-based deployment shim for model serving.
// It uses server-side apply to declaratively manage Deployment resources.
type K8sShimlet struct {
	client *k8s.K8sClient
}

// ID returns the unique identifier for this shimlet.
func (k *K8sShimlet) ID() string { return "k8s" }

// InitWithConfig initializes the K8sShimlet with configuration from the given path.
// It reads the K8s-specific config and establishes a connection to the cluster.
func (k *K8sShimlet) InitWithConfig(confPath string) error {
	k8sCfg, err := config.GetConfFromFileDir[cfg.K8sConfig](confPath)
	if err != nil {
		return err
	}
	client, err := k8s.NewK8sClient(k8sCfg)
	if err != nil {
		return errors.New("failed to initialize K8s client")
	}
	if k == nil {
		return nil
	}
	k.client = client
	return nil
}

// Apply deploys a model server using a Kubernetes Deployment via Server-Side Apply.
// Key fixes and features:
//   - Strips 'local:' prefix from model path (vLLM does not accept it)
//   - Ensures modelDirPath refers to a directory, not a file
//   - Correctly sets HostPath volume type to Directory
//   - Mounts the model volume at the same path used in --model and MODEL env
//
// Returns a success message with exposed port, or an error if deployment fails.
func (k *K8sShimlet) Apply(deploySpec *dto.RequirementSpec) error {
	// Generate deployment and container names
	deploymentName := utils.ModelNameToDeploymentName(deploySpec.ModelName) + "-" + deploySpec.ServiceId
	mainContainerName := utils.ModelNameToDeploymentName(deploySpec.ModelName)
	imageName := "artifacts.iflytek.com/docker-private/aiaas/vllm-openai:v0.4.2"
	modelDirPath := deploySpec.ModelFileDir // Use mapped model path from pipeline

	// Validate model path is provided
	if modelDirPath == "" {
		return errors.New("model path cannot be empty; please provide a valid model name")
	}

	// If the path points to a model file, extract its parent directory
	if strings.HasSuffix(strings.ToLower(modelDirPath), ".bin") ||
		strings.HasSuffix(strings.ToLower(modelDirPath), ".safetensors") ||
		strings.HasSuffix(strings.ToLower(modelDirPath), ".pt") ||
		strings.HasSuffix(strings.ToLower(modelDirPath), ".gguf") {
		modelDirPath = filepath.Dir(modelDirPath)
	}

	// Final validation of resolved model directory path
	if modelDirPath == "" || modelDirPath == "." || modelDirPath == "/" {
		return errors.New("resolved model path is invalid")
	}

	// Initialize container configuration
	container := &corev1apply.ContainerApplyConfiguration{}
	container.WithName(mainContainerName)
	container.WithImage(imageName)
	container.WithImagePullPolicy(corev1.PullIfNotPresent)

	// Configure resource requirements if specified
	if deploySpec.ResourceRequirements != nil {
		resources := &corev1apply.ResourceRequirementsApplyConfiguration{}
		resources.WithRequests(corev1.ResourceList{})
		resources.WithLimits(corev1.ResourceList{})

		if deploySpec.ResourceRequirements.AcceleratorType != "" && deploySpec.ResourceRequirements.AcceleratorCount > 0 {
			limits := corev1.ResourceList{}
			acceleratorResource := corev1.ResourceName(deploySpec.ResourceRequirements.AcceleratorType)
			limits[acceleratorResource] = resource.MustParse(fmt.Sprintf("%d", deploySpec.ResourceRequirements.AcceleratorCount))
			resources.WithLimits(limits)
		}

		container.WithResources(resources)
	}

	// Allocate random NodePort in range 30000–32767
	randomPort := rand.Int31n(2768) + 30000
	portStr := fmt.Sprintf("%d", randomPort)

	// Define environment variables for the container
	envVars := []*corev1apply.EnvVarApplyConfiguration{
		{
			Name:  ptr("MODEL"),
			Value: &modelDirPath, // Model root directory (no 'local:' prefix)
		},
		{
			Name:  ptr("SERVING_ENGINE"),
			Value: ptr("openai"),
		},
		{
			Name:  ptr("PORT"),
			Value: &portStr,
		},
		{
			Name:  ptr("TRANSFORMERS_OFFLINE"),
			Value: ptr("1"), // Enforce offline mode to prevent Hugging Face downloads
		},
		{
			Name:  ptr("HF_HOME"),
			Value: ptr("/tmp"), // Avoid permission issues with default HF cache location
		},
		{
			Name:  ptr("SERVICE_ID"),
			Value: ptr(deploySpec.ServiceId), // Persist serviceId in environment variable
		},
	}

	// Append custom environment variables from deployment spec
	for _, env := range deploySpec.Env {
		envVar := &corev1apply.EnvVarApplyConfiguration{}
		envVar.WithName(env.Key)
		envVar.WithValue(env.Value)
		envVars = append(envVars, envVar)
	}
	container.WithEnv(envVars...)

	// Expose HTTP port on the container
	container.WithPorts(
		corev1apply.ContainerPort().
			WithName("http").
			WithContainerPort(randomPort),
	)

	// Set command-line arguments for vLLM OpenAI API server
	container.WithArgs(
		"--host=0.0.0.0",
		"--port="+portStr,
		"--model="+modelDirPath, // Model path must match volume mount
		"--dtype=auto",
		"--trust-remote-code", // Required for models like Qwen
	)

	// Build Deployment object using Apply Configuration pattern
	deploymentApply := &appsv1apply.DeploymentApplyConfiguration{}
	deploymentApply.WithAPIVersion("apps/v1")
	deploymentApply.WithKind("Deployment")
	deploymentApply.WithName(deploymentName)
	deploymentApply.WithNamespace("default")
	deploymentApply.WithLabels(map[string]string{
		"app":        deploySpec.ServiceId,
		"managed-by": "astron-xmod-shim",
	})
	deploymentApply.WithAnnotations(map[string]string{
		"astron-xmod-shim/service-id": deploySpec.ServiceId,
		"astron-xmod-shim/model-name": deploySpec.ModelName,
	})

	// Configure Deployment spec
	spec := &appsv1apply.DeploymentSpecApplyConfiguration{}
	spec.WithReplicas(int32(deploySpec.ReplicaCount))

	// Define label selector for Pod matching
	selector := &metav1apply.LabelSelectorApplyConfiguration{}
	selector.WithMatchLabels(map[string]string{"app": deploySpec.ServiceId})
	spec.WithSelector(selector)

	// Configure Pod template
	template := &corev1apply.PodTemplateSpecApplyConfiguration{}
	template.WithLabels(map[string]string{"app": deploySpec.ServiceId})

	// Configure Pod specification
	podSpec := &corev1apply.PodSpecApplyConfiguration{}
	podSpec.WithHostNetwork(true) // Use host network for direct port exposure

	// Set nodeSelector as a local variable
	// This can be modified to match specific node requirements
	nodeSelector := map[string]string{}
	nodeSelector["kubernetes.io/hostname"] = "dx-l20-10.246.53.166.maas.cn"
	// Example: To schedule on nodes with GPU label
	// nodeSelector["nvidia.com/gpu.present"] = "true"
	// Enable nodeSelector if it contains any key-value pairs
	if len(nodeSelector) > 0 {
		podSpec.WithNodeSelector(nodeSelector)
	}

	// Tolerate all taints to allow scheduling on dedicated GPU nodes
	podSpec.WithTolerations(
		corev1apply.Toleration().
			WithKey("").
			WithOperator(corev1.TolerationOpExists),
	)

	// Mount host model directory into the container using HostPath
	podSpec.WithVolumes(
		corev1apply.Volume().
			WithName("models").
			WithHostPath(
				corev1apply.HostPathVolumeSource().
					WithPath(modelDirPath).             // Host machine path
					WithType(corev1.HostPathDirectory), // Ensure it's treated as a directory
			),
	)

	// Mount the volume inside the container at the exact model path
	container.WithVolumeMounts(
		corev1apply.VolumeMount().
			WithName("models").
			WithMountPath(modelDirPath), // Must match --model argument
	)

	// Attach container to Pod spec
	podSpec.WithContainers(container)

	// Attach Pod spec to template
	template.WithSpec(podSpec)

	// Attach template to Deployment spec
	spec.WithTemplate(template)
	deploymentApply.WithSpec(spec)

	// Perform Server-Side Apply to create or update the Deployment
	result, err := k.client.GetClientSet().AppsV1().Deployments("default").Apply(
		context.Background(),
		deploymentApply,
		metav1.ApplyOptions{FieldManager: "astron-xmod-shim", Force: true},
	)
	if err != nil {
		return fmt.Errorf("failed to deploy application: %w", err)
	}
	log.Info("Deployment %s/%s succeeded with hostNetwork on port %d", result.Namespace, result.Name, randomPort)
	return nil
}

// ptr creates a pointer to a string value (helper for ApplyConfigurations).
func ptr(s string) *string { return &s }

// Delete removes deployed resources associated with the given resourceId.
// In our implementation, resourceId corresponds to serviceId, which is used to find
// and delete all Kubernetes Deployments labeled with this serviceId.
func (k *K8sShimlet) Delete(resourceId string) error {
	if k.client == nil {
		return errors.New("K8s client is not initialized")
	}

	// Use label selector to find deployments with the given serviceId
	labelSelector := labels.Set{"app": resourceId}.AsSelector().String()
	opts := metav1.ListOptions{LabelSelector: labelSelector}

	// List deployments with the specified serviceId
	deployments, err := k.client.ListDeployments("default", opts)
	if err != nil {
		return fmt.Errorf("failed to list deployments for service %s: %w", resourceId, err)
	}

	// Delete each found deployment
	for _, deployment := range deployments {
		// Delete the deployment using Kubernetes API
		err = k.client.GetClientSet().AppsV1().Deployments(deployment.Namespace).Delete(
			context.Background(),
			deployment.Name,
			metav1.DeleteOptions{},
		)
		if err != nil {
			// Continue deleting other deployments even if one fails
			log.Error("Failed to delete deployment %s/%s: %v", deployment.Namespace, deployment.Name, err)
		} else {
			log.Info("Successfully deleted deployment %s/%s", deployment.Namespace, deployment.Name)
		}
	}

	// If no deployments were found, consider it a success (already deleted)
	if len(deployments) == 0 {
		log.Info("No deployments found for service %s", resourceId)
	}

	return nil
}

// Status retrieves the current status of a deployed resource based on Kubernetes deployment state.
// Status retrieves the current status of a deployed resource and its endpoint.
func (k *K8sShimlet) Status(resourceId string) (*dto.RuntimeStatus, error) {
	if k.client == nil {
		return nil, errors.New("K8s client is not initialized")
	}

	// Use label selector to find deployments with the given resourceId (serviceId)
	labelSelector := labels.Set{"app": resourceId}.AsSelector().String()
	opts := metav1.ListOptions{LabelSelector: labelSelector}

	// List deployments with the specified resourceId
	deployments, err := k.client.ListDeployments("default", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments for resource %s: %w", resourceId, err)
	}

	// If no deployments found, return terminated status
	if len(deployments) == 0 {
		return &dto.RuntimeStatus{
			DeploySpec: dto.RequirementSpec{ServiceId: resourceId},
			Status:     dto.PhaseUnknown,
		}, nil
	}

	// Get the first deployment (assuming one per serviceId)
	deployment := deployments[0]

	// Determine deployment status
	var phase dto.DeployPhase
	switch {
	case deployment.Status.Replicas == 0:
		phase = dto.PhaseTerminating
	case deployment.Status.UnavailableReplicas > 0:
		phase = dto.PhaseFailed
	case deployment.Status.AvailableReplicas == deployment.Status.Replicas:
		phase = dto.PhaseRunning
	default:
		phase = dto.PhasePending
	}

	// Extract model name and path from annotations or labels
	modelName := "unknown"
	modelPath := "unknown"

	if val, ok := deployment.Annotations["astron-xmod-shim/model-name"]; ok {
		modelName = val
	}
	if val, ok := deployment.Annotations["astron-xmod-shim/model-path"]; ok {
		modelPath = val
	}

	// Extract replica count
	replicaCount := int(*deployment.Spec.Replicas)

	// 🌟 新增：从 PodTemplate 中提取容器端口（即 NodePort）
	var nodePort int32 = 0
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		for _, c := range deployment.Spec.Template.Spec.Containers {
			if c.Ports != nil {
				for _, p := range c.Ports {
					if p.Name == "http" {
						nodePort = p.ContainerPort
						break
					}
				}
			}
			if nodePort != 0 {
				break
			}
		}
	}

	// 新增：获取任一运行中的 Pod 的 Node IP
	var nodeIP string
	if nodePort != 0 {
		// 列出该 Deployment 的所有 Pod
		podListOptions := metav1.ListOptions{
			LabelSelector: labels.Set{"app": resourceId}.AsSelector().String(),
		}
		pods, err := k.client.GetClientSet().CoreV1().Pods("default").List(context.Background(), podListOptions)
		if err != nil {
			log.Warn("Failed to list pods for deployment %s: %v", deployment.Name, err)
		} else {
			for _, pod := range pods.Items {
				if pod.Spec.NodeName != "" && pod.Status.Phase == corev1.PodRunning {
					// 获取 Node 对象
					node, err := k.client.GetClientSet().CoreV1().Nodes().Get(context.Background(), pod.Spec.NodeName, metav1.GetOptions{})
					if err != nil {
						continue
					}
					// 查找 InternalIP
					for _, addr := range node.Status.Addresses {
						if addr.Type == corev1.NodeInternalIP {
							nodeIP = addr.Address
							break
						}
					}
					if nodeIP != "" {
						break // 使用第一个运行中 Pod 的节点 IP
					}
				}
			}
		}
	}

	// 🌟 构造 endpoint
	var endpoint string
	if nodeIP != "" && nodePort != 0 {
		endpoint = fmt.Sprintf("http://%s:%d", nodeIP, nodePort)
	}

	// 从Deployment中提取ResourceRequirements信息
	var resourceRequirements *dto.ResourceRequirements
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := deployment.Spec.Template.Spec.Containers[0]
		if len(container.Resources.Limits) > 0 {
			for resourceName, quantity := range container.Resources.Limits {
				// 检查是否是GPU资源
				if strings.Contains(string(resourceName), "gpu") || strings.Contains(string(resourceName), "nvidia") {
					resourceRequirements = &dto.ResourceRequirements{
						AcceleratorType:  string(resourceName),
						AcceleratorCount: int(quantity.Value()),
					}
					break
				}
			}
		}
	}

	// 从Deployment的环境变量中提取ContextLength和Env信息
	var contextLength int
	var envVars []dto.Env
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := deployment.Spec.Template.Spec.Containers[0]
		for _, envVar := range container.Env {
			switch envVar.Name {
			case "CONTEXT_LENGTH":
				if val, err := strconv.Atoi(envVar.Value); err == nil {
					contextLength = val
				}
			default:
				envVars = append(envVars, dto.Env{
					Key:   envVar.Name,
					Value: envVar.Value,
				})
			}
		}
	}

	// 从Deployment注解中提取GoalSetName和ShimletName
	goalSetName := "opensource-llm-deploy" // 默认值
	shimletName := "k8s"                   // 默认值

	if val, ok := deployment.Annotations["astron-xmod-shim/goal-set-name"]; ok {
		goalSetName = val
	}
	if val, ok := deployment.Annotations["astron-xmod-shim/shimlet-name"]; ok {
		shimletName = val
	}

	// Build deploy spec
	spec := dto.RequirementSpec{
		ServiceId:            resourceId,
		ModelName:            modelName,
		ModelFileDir:         modelPath,
		ResourceRequirements: resourceRequirements,
		ReplicaCount:         replicaCount,
		ContextLength:        contextLength,
		Env:                  envVars,
		GoalSetName:          goalSetName,
		ShimletName:          shimletName,
	}

	return &dto.RuntimeStatus{
		DeploySpec: spec,
		Status:     phase,
		EndPoint:   endpoint, // ✅ 返回 endpoint
	}, nil
}

// Description returns a brief description of the shimlet.
func (k *K8sShimlet) Description() string { return "k8s shimlet" }

// ListDeployedServices 获取所有已部署的服务列表
// 这个方法查询Kubernetes集群中所有由astron-xmod-shim管理的部署，并提取对应的serviceId
func (k *K8sShimlet) ListDeployedServices() ([]string, error) {
	if k.client == nil {
		return []string{}, fmt.Errorf("k8s client is not initialized")
	}

	// 准备ListOptions，筛选由astron-xmod-shim管理的部署
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set{"managed-by": "astron-xmod-shim"}.AsSelector().String(),
	}

	// 调用ListDeployments方法获取所有由astron-xmod-shim管理的部署
	deployments, err := k.client.ListDeployments("default", listOptions)
	if err != nil {
		return []string{}, fmt.Errorf("failed to list deployments: %w", err)
	}

	// 从部署中提取serviceId
	var serviceIDs []string
	for _, deployment := range deployments {
		// 检查deployment是否有astron-xmod-shim/service-id注解
		if serviceID, exists := deployment.Annotations["astron-xmod-shim/service-id"]; exists && serviceID != "" {
			serviceIDs = append(serviceIDs, serviceID)
		} else {
			// 尝试从标签中获取serviceId
			if appLabel, exists := deployment.Labels["app"]; exists && appLabel != "" {
				serviceIDs = append(serviceIDs, appLabel)
			}
		}
	}

	return serviceIDs, nil
}
