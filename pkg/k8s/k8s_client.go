package k8s

import (
	"astron-xmod-shim/pkg/log"
	"context"
	"errors"
	"fmt"
	"regexp"
	stdruntime "runtime"
	"strconv"
	"time"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors" // Critical: Must import this package
	"k8s.io/apimachinery/pkg/labels"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	// 引入配置结构体
	config "astron-xmod-shim/internal/dto/config"
)

// K8sClient 通用K8s客户端，直接包含所有Informer组件和客户端实例
type K8sClient struct {
	client *kubernetes.Clientset // 原生clientset

	// Informer相关组件
	podInformer    cache.SharedIndexInformer                    // Pod Informer
	deployInformer cache.SharedIndexInformer                    // Deployment Informer
	cmInformer     cache.SharedIndexInformer                    // CM Informer
	podLister      cache.GenericLister                          // Pod缓存查询器
	deployLister   cache.GenericLister                          // Deployment缓存查询器
	nodeInformer   cache.SharedIndexInformer                    // 节点Informer（新增）
	nodeLister     cache.GenericLister                          // 节点缓存查询器（新增）
	cmLister       cache.GenericLister                          // CM缓存查询器
	stopper        chan struct{}                                // 用于停止Informer的信号通道
	queue          workqueue.TypedRateLimitingInterface[string] // 泛型事件队列
}

// NewK8sClient 初始化K8s客户端（直接初始化所有组件）
func NewK8sClient(cfg *config.K8sConfig) (*K8sClient, error) {
	if cfg == nil {
		return nil, errors.New("K8s配置不能为空")
	}

	// 1. 构建REST配置（集群内/外兼容）
	restCfg, err := buildRestConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("构建REST配置失败: %w", err)
	}

	// 2. 应用客户端配置（QPS、超时等）
	applyClientConfig(restCfg, cfg)

	// 3. 创建原生clientset
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("创建K8s clientset失败: %w", err)
	}

	// 4. 初始化客户端实例（直接初始化所有字段）
	client := &K8sClient{
		client:  clientset,
		stopper: make(chan struct{}), // 初始化停止信号通道
	}

	// 5. 初始化Pod Informer及Lister（使用NewFilteredListWatchFromClient替代手动List/Watch）
	client.podInformer = cache.NewSharedIndexInformer(
		// 使用官方推荐方法创建ListWatch，自动处理List/Watch逻辑
		cache.NewFilteredListWatchFromClient(
			clientset.CoreV1().RESTClient(), // 传入Pod资源的RESTClient
			"pods",                          // 资源名称（字符串）
			metav1.NamespaceAll,             // 监听所有命名空间
			func(opts *metav1.ListOptions) { // 全局筛选器（无筛选可留空）
			},
		),
		&corev1.Pod{}, // 资源对象类型
		5*time.Minute, // 缓存重同步间隔
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, // 命名空间索引
	)
	client.podLister = cache.NewGenericLister(
		client.podInformer.GetIndexer(),
		corev1.SchemeGroupVersion.WithResource("pods").GroupResource(),
	)

	// 6. 初始化Deployment Informer及Lister（同样使用推荐方法）
	client.deployInformer = cache.NewSharedIndexInformer(
		cache.NewFilteredListWatchFromClient(
			clientset.AppsV1().RESTClient(), // 传入Deployment资源的RESTClient
			"deployments",                   // 资源名称（字符串）
			metav1.NamespaceAll,
			func(opts *metav1.ListOptions) {}, // 无筛选
		),
		&appsv1.Deployment{},
		5*time.Minute,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	client.deployLister = cache.NewGenericLister(
		client.deployInformer.GetIndexer(),
		appsv1.SchemeGroupVersion.WithResource("deployments").GroupResource(),
	)

	// 7. 初始化CM Informer及Lister（同样使用推荐方法）
	client.cmInformer = cache.NewSharedIndexInformer(
		cache.NewFilteredListWatchFromClient(
			clientset.CoreV1().RESTClient(), // 传入Deployment资源的RESTClient
			"configmaps",                    // 资源名称（字符串）
			metav1.NamespaceAll,
			func(opts *metav1.ListOptions) {}, // 无筛选
		),
		&corev1.ConfigMap{},
		5*time.Minute,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	client.cmLister = cache.NewGenericLister(
		client.cmInformer.GetIndexer(),
		corev1.SchemeGroupVersion.WithResource("configmaps").GroupResource(),
	)

	//8. 新增：初始化Node Informer及Lister
	client.nodeInformer = cache.NewSharedIndexInformer(
		cache.NewFilteredListWatchFromClient(
			clientset.CoreV1().RESTClient(), // 节点属于corev1资源
			"nodes",                         // 资源名称（节点是集群级资源，无命名空间）
			metav1.NamespaceAll,
			func(opts *metav1.ListOptions) {}, // 无筛选
		),
		&corev1.Node{},   // 资源对象类型
		5*time.Minute,    // 缓存重同步间隔
		cache.Indexers{}, // 节点无需命名空间索引（集群级资源）
	)
	client.nodeLister = cache.NewGenericLister(
		client.nodeInformer.GetIndexer(),
		corev1.SchemeGroupVersion.WithResource("nodes").GroupResource(),
	)

	// 9. 初始化泛型事件队列（适配client-go v0.33.3）
	client.queue = workqueue.NewTypedRateLimitingQueueWithConfig(
		workqueue.DefaultTypedControllerRateLimiter[string](), // 默认限流策略
		workqueue.TypedRateLimitingQueueConfig[string]{
			Name: "k8s-resource-event-queue", // 队列名称（用于监控和日志）
		},
	)

	// 10. 注册事件处理器
	client.registerEventHandlers()

	// 11. 启动Informer和事件处理
	client.startInformerSystem()

	// 12. 等待缓存同步完成
	if !cache.WaitForCacheSync(
		client.stopper,
		client.podInformer.HasSynced,
		client.deployInformer.HasSynced,
		client.cmInformer.HasSynced,
		client.nodeInformer.HasSynced,
	) {
		return nil, errors.New("informer缓存同步超时")
	}

	return client, nil
}

// buildRestConfig 根据配置选择集群内/外配置
func buildRestConfig(cfg *config.K8sConfig) (*rest.Config, error) {
	if cfg.Kubeconfig != "" {
		// 集群外：使用指定kubeconfig
		return clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}
	// 集群内：使用serviceaccount
	return rest.InClusterConfig()
}

// applyClientConfig 应用客户端性能配置
func applyClientConfig(restCfg *rest.Config, cfg *config.K8sConfig) {
	// QPS限制（默认10）
	restCfg.QPS = float32(cfg.QPS)
	if restCfg.QPS <= 0 {
		restCfg.QPS = 10
	}

	// 突发流量限制（默认20）
	restCfg.Burst = cfg.Burst
	if restCfg.Burst <= 0 {
		restCfg.Burst = 20
	}

	// 超时时间（默认30s）
	defaultTimeout := 30 * time.Second
	restCfg.Timeout = defaultTimeout
	if cfg.Timeout > 0 {
		if timeout, err := time.ParseDuration(strconv.FormatInt(cfg.Timeout, 10) + "s"); err == nil {
			restCfg.Timeout = timeout
		}
	}
}

// registerEventHandlers 注册事件处理器（实例方法，直接访问组件）
func (c *K8sClient) registerEventHandlers() {
	// Pod事件处理
	_, _ = c.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("pod", "add", obj))
		},
		UpdateFunc: func(old, new interface{}) {
			// 忽略资源版本未变化的更新
			oldPod, oldOk := old.(*corev1.Pod)
			newPod, newOk := new.(*corev1.Pod)
			if oldOk && newOk && oldPod.ResourceVersion == newPod.ResourceVersion {
				return
			}
			c.queue.AddRateLimited(eventKey("pod", "update", new))
		},
		DeleteFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("pod", "delete", obj))
		},
	})

	// Deployment事件处理
	_, _ = c.deployInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("deploy", "add", obj))
		},
		UpdateFunc: func(old, new interface{}) {
			// 忽略资源版本未变化的更新
			oldDeploy, oldOk := old.(*appsv1.Deployment)
			newDeploy, newOk := new.(*appsv1.Deployment)
			if oldOk && newOk && oldDeploy.ResourceVersion == newDeploy.ResourceVersion {
				return
			}
			c.queue.AddRateLimited(eventKey("deploy", "update", new))
		},
		DeleteFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("deploy", "delete", obj))
		},
	})

	// CM事件处理
	_, _ = c.cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("CM", "add", obj))
		},
		UpdateFunc: func(old, new interface{}) {
			// 忽略资源版本未变化的更新
			oldCM, oldOk := old.(*corev1.ConfigMap)
			newCM, newOk := new.(*corev1.ConfigMap)
			if oldOk && newOk && oldCM.ResourceVersion == newCM.ResourceVersion {
				return
			}
			c.queue.AddRateLimited(eventKey("CM", "update", new))
		},
		DeleteFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("CM", "delete", obj))
		},
	})

	// node事件处理
	_, _ = c.nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("node", "add", obj))
		},
		UpdateFunc: func(old, new interface{}) {
			oldNode, oldOk := old.(*corev1.Node)
			newNode, newOk := new.(*corev1.Node)
			if oldOk && newOk && oldNode.ResourceVersion == newNode.ResourceVersion {
				return
			}
			c.queue.AddRateLimited(eventKey("node", "update", new))
		},
		DeleteFunc: func(obj interface{}) {
			c.queue.AddRateLimited(eventKey("node", "delete", obj))
		},
	})

}

// startInformerSystem 启动Informer和事件处理协程
func (c *K8sClient) startInformerSystem() {
	// 启动Informer（独立协程）
	go c.podInformer.Run(c.stopper)
	go c.deployInformer.Run(c.stopper)
	go c.cmInformer.Run(c.stopper)
	go c.nodeInformer.Run(c.stopper)
	// 启动事件处理协程
	go c.processEvents()
}

func (c *K8sClient) processEvents() {
	defer stdruntime.KeepAlive(c.queue)
	for {
		key, shutdown := c.queue.Get()
		if shutdown {
			return
		}
		// 用匿名函数包裹defer 防止内存泄漏
		func() {
			defer c.queue.Done(key)
			//fmt.Printf("处理事件: %s\n", key)
			c.queue.Forget(key)
		}() // 立即执行匿名函数
	}
}

// Stop 停止客户端及所有Informer
func (c *K8sClient) Stop() {
	select {
	case <-c.stopper:
		// 已停止，避免重复关闭
	default:
		close(c.stopper)
	}
	c.queue.ShutDown()
}

// ListPods 从缓存查询指定命名空间的Pod（支持标签筛选）
func (c *K8sClient) ListPods(namespace string, opts metav1.ListOptions) ([]*corev1.Pod, error) {
	selector := labels.Everything()
	if opts.LabelSelector != "" {
		var err error
		selector, err = labels.Parse(opts.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("解析标签选择器失败: %w", err)
		}
	}

	// 从Lister查询缓存
	objs, err := c.podLister.ByNamespace(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("查询Pod缓存失败: %w", err)
	}

	// 类型转换
	pods := make([]*corev1.Pod, 0, len(objs))
	for _, obj := range objs {
		if pod, ok := obj.(*corev1.Pod); ok {
			pods = append(pods, pod)
		}
	}
	return pods, nil
}

// ListDeployments 从缓存查询指定命名空间的Deployment
func (c *K8sClient) ListDeployments(namespace string, opts metav1.ListOptions) ([]*appsv1.Deployment, error) {
	selector := labels.Everything()
	if opts.LabelSelector != "" {
		var err error
		selector, err = labels.Parse(opts.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("解析标签选择器失败: %w", err)
		}
	}

	// 从Lister查询缓存
	objs, err := c.deployLister.ByNamespace(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("查询Deployment缓存失败: %w", err)
	}

	// 类型转换
	deploys := make([]*appsv1.Deployment, 0, len(objs))
	for _, obj := range objs {
		if deploy, ok := obj.(*appsv1.Deployment); ok {
			deploys = append(deploys, deploy)
		}
	}
	return deploys, nil
}

// GetClientSet 暴露原生clientset（用于直接调用K8s API）
func (c *K8sClient) GetClientSet() *kubernetes.Clientset {
	return c.client
}

// eventKey 生成事件唯一标识（格式：资源类型/命名空间/事件类型/资源名称）
func eventKey(resource, eventType string, obj interface{}) string {
	if metaObj, ok := obj.(metav1.Object); ok {
		return fmt.Sprintf("%s/%s/%s/%s", resource, metaObj.GetNamespace(), eventType, metaObj.GetName())
	}
	return fmt.Sprintf("%s/%s/unknown", resource, eventType)
}

func (c *K8sClient) UpsertConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	// 1. 验证参数合法性
	if err := validateConfigMapName(name); err != nil {
		return nil, fmt.Errorf("ConfigMap名称不合法: %w", err)
	}
	if namespace == "" {
		return nil, errors.New("命名空间不能为空")
	}
	// TODO 对接 llm cm

	// 3. 构建目标ConfigMap对象
	targetCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        "",
				"resource":   "llm-cm",
				"managed-by": "astron-xmod-shim",
			},
			Annotations: map[string]string{
				"last-updated": time.Now().Format(time.RFC3339),
			},
		},
		Data: map[string]string{
			"llm.json": string("strategyJSON"),
		},
	}

	// 4. 尝试获取现有ConfigMap
	existingCM, err := c.client.CoreV1().ConfigMaps(namespace).Get(
		context.Background(),
		name,
		metav1.GetOptions{},
	)

	// 5. 处理存在/不存在的情况（使用errors.IsNotFound判断）
	if err != nil {
		// 资源不存在，创建新的
		if k8s_errors.IsNotFound(err) {
			return c.client.CoreV1().ConfigMaps(namespace).Create(
				context.Background(),
				targetCM,
				metav1.CreateOptions{},
			)
		}
		// 其他错误（如权限不足）
		return nil, fmt.Errorf("查询ConfigMap失败: %w", err)
	}

	// 6. 资源已存在，更新（保留ResourceVersion实现乐观锁）
	targetCM.ResourceVersion = existingCM.ResourceVersion
	// Preserve existing labels (avoid overwriting other labels)
	for k, v := range existingCM.Labels {
		if _, ok := targetCM.Labels[k]; !ok {
			targetCM.Labels[k] = v
		}
	}
	// Preserve existing annotations (only update last-updated)
	for k, v := range existingCM.Annotations {
		if k != "last-updated" {
			targetCM.Annotations[k] = v
		}
	}

	// Execute update
	return c.client.CoreV1().ConfigMaps(namespace).Update(
		context.Background(),
		targetCM,
		metav1.UpdateOptions{},
	)
}

// validateConfigMapName Validate ConfigMap name legality
func validateConfigMapName(name string) error {
	if len(name) == 0 {
		return errors.New("name cannot be empty")
	}
	if len(name) > 63 {
		return errors.New("name length cannot exceed 63 characters")
	}
	if !regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`).MatchString(name) {
		return errors.New("name can only contain lowercase letters, numbers, hyphens (-), and dots (.), and cannot start or end with a hyphen")
	}
	return nil
}

// GetConfigMapFromCache Query ConfigMap from cache (fixed)
func (c *K8sClient) GetConfigMapFromCache(namespace, name string) (*corev1.ConfigMap, error) {
	// 1. Get namespace-specific Lister through GenericLister
	nsLister := c.cmLister.ByNamespace(namespace)

	// 2. Query specified ConfigMap from namespace Lister
	obj, err := nsLister.Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to query ConfigMap from cache: %w", err)
	}

	// 3. Type conversion to *corev1.ConfigMap
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil, errors.New("resource in cache is not of type ConfigMap")
	}

	return cm, nil
}

func (c *K8sClient) DeleteConfigMap(ns string, name string) error {
	return c.client.CoreV1().ConfigMaps(ns).Delete(
		context.Background(),
		name,
		metav1.DeleteOptions{},
	)
}

// ListNodesByLabelFromCache Query nodes with specified labels from informer cache
func (c *K8sClient) ListNodesByLabelFromCache(labelSelector string) ([]*corev1.Node, error) {
	// Parse label selector
	selector := labels.Everything()
	if labelSelector != "" {
		var err error
		selector, err = labels.Parse(labelSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to parse label selector: %w", err)
		}
	}

	// Query from node Lister cache (nodes are cluster-level resources, no namespace needed)
	objs, err := c.nodeLister.List(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes from cache: %w", err)
	}

	// Type conversion to *corev1.Node
	nodes := make([]*corev1.Node, 0, len(objs))
	for _, obj := range objs {
		if node, ok := obj.(*corev1.Node); ok {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

// NodesHaveTaint Determine if node list contains specified taint
// taintKey: Taint key (e.g., "node-role.kubernetes.io/master")
// taintEffect: Taint effect (e.g., "NoSchedule", empty means no effect check)
func (c *K8sClient) NodesHaveTaint(nodes []*corev1.Node, taintKey, taintEffect string) map[string]bool {
	result := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		result[node.Name] = false // 默认不包含
		// 遍历节点的污点列表
		for _, taint := range node.Spec.Taints {
			// 匹配污点键，若指定效果则同时匹配
			if taint.Key == taintKey && (taintEffect == "" || string(taint.Effect) == taintEffect) {
				result[node.Name] = true
				break
			}
		}
	}
	return result
}

// AddTaintsToNodes Add taints to specified node list (idempotent handling)
func (c *K8sClient) AddTaintsToNodes(nodes []*corev1.Node, taintKey, taintValue, taintEffect string) error {
	if len(nodes) == 0 {
		return errors.New("node list cannot be empty")
	}
	if taintKey == "" || taintEffect == "" {
		return errors.New("taint key and effect cannot be empty")
	}

	// Validate taint effect legality
	validEffects := map[corev1.TaintEffect]bool{
		corev1.TaintEffectNoSchedule:       true,
		corev1.TaintEffectNoExecute:        true,
		corev1.TaintEffectPreferNoSchedule: true,
	}
	effect := corev1.TaintEffect(taintEffect)
	if !validEffects[effect] {
		return fmt.Errorf("unsupported taint effect: %s", taintEffect)
	}

	targetTaint := corev1.Taint{
		Key:    taintKey,
		Value:  taintValue,
		Effect: effect,
	}

	for _, node := range nodes {
		// Check if same taint already exists
		hasTaint := false
		for _, t := range node.Spec.Taints {
			if t.Key == targetTaint.Key && t.Effect == targetTaint.Effect {
				// Update if values differ
				if t.Value != targetTaint.Value {
					log.Info("Node %s taint %s:%s has different value, will update", node.Name, taintKey, taintEffect)
				} else {
					hasTaint = true
					break
				}
			}
		}

		if hasTaint {
			log.Info("Node %s already contains taint %s:%s, skipping", node.Name, taintKey, taintEffect)
			continue
		}

		// Deep copy to avoid modifying cache
		nodeCopy := node.DeepCopy()
		nodeCopy.Spec.Taints = append(nodeCopy.Spec.Taints, targetTaint)

		// Execute update
		_, err := c.client.CoreV1().Nodes().Update(context.Background(), nodeCopy, metav1.UpdateOptions{})
		if err != nil {
			log.Error("Failed to add taint to node %s: %v", node.Name, err)
			continue
		}
		log.Info("Successfully added taint %s=%s:%s to node %s", node.Name, taintKey, taintValue, taintEffect)
	}

	return nil
}

// RemoveTaintsFromNodes Remove specified taint from node list
func (c *K8sClient) RemoveTaintsFromNodes(nodes []*corev1.Node, taintKey, taintEffect string) error {
	if len(nodes) == 0 {
		return errors.New("node list cannot be empty")
	}
	if taintKey == "" {
		return errors.New("taint key cannot be empty")
	}

	// Convert taint effect (allow empty value to indicate no effect restriction)
	var effect corev1.TaintEffect
	if taintEffect != "" {
		effect = corev1.TaintEffect(taintEffect)
	}

	for _, node := range nodes {
		// Check if target taint exists
		taintIndex := -1
		for i, t := range node.Spec.Taints {
			if t.Key == taintKey && (taintEffect == "" || t.Effect == effect) {
				taintIndex = i
				break
			}
		}

		if taintIndex == -1 {
			log.Info("Node %s does not have taint %s:%s, skipping", node.Name, taintKey, taintEffect)
			continue
		}

		// Deep copy to avoid modifying cache
		nodeCopy := node.DeepCopy()
		// Remove target taint
		nodeCopy.Spec.Taints = append(
			nodeCopy.Spec.Taints[:taintIndex],
			nodeCopy.Spec.Taints[taintIndex+1:]...,
		)

		// Execute update
		_, err := c.client.CoreV1().Nodes().Update(context.Background(), nodeCopy, metav1.UpdateOptions{})
		if err != nil {
			log.Error("Failed to remove taint from node %s: %v", node.Name, err)
			continue
		}
		log.Info("Successfully removed taint %s:%s from node %s", node.Name, taintKey, taintEffect)
	}

	return nil
}

// ListPendingPodNamesInNamespace Query list of Pod names in Pending state in specified namespace
func (c *K8sClient) ListPendingPodNamesInNamespace(namespace string) ([]string, error) {
	if namespace == "" {
		return nil, errors.New("namespace cannot be empty")
	}

	// Query all Pods in the namespace from cache
	pods, err := c.ListPods(namespace, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to query Pods in namespace %s: %w", namespace, err)
	}

	// Filter Pod names in Pending state
	var pendingPodNames []string
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodPending {
			pendingPodNames = append(pendingPodNames, pod.Name)
		}
	}

	return pendingPodNames, nil
}

// DeletePodsOnNodesInNamespace Delete Pods on specified nodes and namespace
func (c *K8sClient) DeletePodsOnNodesInNamespace(namespace string, nodes []*corev1.Node) error {
	if namespace == "" {
		return errors.New("namespace cannot be empty")
	}
	if len(nodes) == 0 {
		return errors.New("node list cannot be empty")
	}

	// Extract node name map for quick lookup
	nodeNameMap := make(map[string]struct{})
	for _, node := range nodes {
		nodeNameMap[node.Name] = struct{}{}
	}

	// Query all Pods in the namespace from cache
	pods, err := c.ListPods(namespace, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to query Pods in namespace %s: %w", namespace, err)
	}

	// Filter Pods on target nodes and delete them
	for _, pod := range pods {
		// Check if Pod is scheduled to target node
		if pod.Spec.NodeName == "" {
			continue // Skip unscheduled Pods
		}
		if _, exists := nodeNameMap[pod.Spec.NodeName]; !exists {
			continue // Not in target node list, skip
		}

		// 执行删除操作
		err := c.client.CoreV1().Pods(namespace).Delete(
			context.Background(),
			pod.Name,
			metav1.DeleteOptions{},
		)
		if err != nil {
			// Ignore not found errors, log other errors and continue
			if !k8s_errors.IsNotFound(err) {
				log.Error("Failed to delete Pod %s/%s: %v", namespace, pod.Name, err)
				continue
			}
		} else {
			log.Info("Successfully deleted Pod %s/%s", namespace, pod.Name)
		}
	}

	return nil
}
