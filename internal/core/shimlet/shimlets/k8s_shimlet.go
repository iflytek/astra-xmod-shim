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
	"modserv-shim/pkg/log"
	"modserv-shim/pkg/utils"

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
func (k *K8sShimlet) Apply(deploySpec *dto.DeploySpec) (string, error) {
	// Generate deployment and container names
	deploymentName := utils.ModelNameToDeploymentName(deploySpec.ModelName) + "-" + deploySpec.ServiceId
	mainContainerName := utils.ModelNameToDeploymentName(deploySpec.ModelName)
	imageName := "artifacts.iflytek.com/docker-private/aiaas/vllm-openai:v0.4.2"
	modelDirPath := deploySpec.ModelFileDir // Use mapped model path from pipeline

	// Validate model path is provided
	if modelDirPath == "" {
		return "", errors.New("model path cannot be empty; please provide a valid model name")
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
		return "", errors.New("resolved model path is invalid")
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
		envVar.WithValue(env.Val)
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
		"managed-by": "modserv-shim",
	})
	deploymentApply.WithAnnotations(map[string]string{
		"modserv-shim/service-id": deploySpec.ServiceId,
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
		metav1.ApplyOptions{FieldManager: "modserv-shim", Force: true},
	)
	if err != nil {
		return "", fmt.Errorf("failed to deploy application: %w", err)
	}

	return fmt.Sprintf("Deployment %s/%s succeeded with hostNetwork on port %d", result.Namespace, result.Name, randomPort), nil
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

// Status retrieves the current status of a deployed resource (not implemented).
func (k *K8sShimlet) Status(resourceId string) (*dto.DeployStatus, error) { return nil, nil }

// Description returns a brief description of the shimlet.
func (k *K8sShimlet) Description() string { return "k8s shimlet" }