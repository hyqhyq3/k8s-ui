package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// K8sService Kubernetes 业务逻辑
type K8sService struct {
	client        *kubernetes.Clientset
	discoveryClient discovery.DiscoveryInterface
}

// NewK8sService 创建 K8s 服务
func NewK8sService(client *kubernetes.Clientset) *K8sService {
	return &K8sService{
		client:        client,
		discoveryClient: client.Discovery(),
	}
}

// NamespaceInfo namespace 信息
type NamespaceInfo struct {
	Name   string              `json:"name"`
	Status corev1.NamespacePhase `json:"status"`
	Age    string              `json:"age"`
	Labels map[string]string   `json:"labels"`
}

// PodInfo pod 信息
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Restarts  int32  `json:"restarts"`
	Node      string `json:"node"`
	Age       string `json:"age"`
	IP        string `json:"ip"`
}

// DeploymentInfo deployment 信息
type DeploymentInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Replicas  int32             `json:"replicas"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Images    []string          `json:"images"`
	Labels    map[string]string `json:"labels"`
}

// StatefulSetInfo statefulset 信息
type StatefulSetInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Replicas  int32             `json:"replicas"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Images    []string          `json:"images"`
	Labels    map[string]string `json:"labels"`
}

// DaemonSetInfo daemonset 信息
type DaemonSetInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Desired   int32             `json:"desired"`
	Ready     int32             `json:"ready"`
	Age       string            `json:"age"`
	Images    []string          `json:"images"`
	Labels    map[string]string `json:"labels"`
}

// ConfigMapInfo configmap 信息
type ConfigMapInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Keys      []string          `json:"keys"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// SecretInfo secret 信息
type SecretInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// PersistentVolumeInfo persistentvolume 信息
type PersistentVolumeInfo struct {
	Name          string   `json:"name"`
	Capacity      string   `json:"capacity"`
	AccessModes   []string `json:"accessModes"`
	ReclaimPolicy string   `json:"reclaimPolicy"`
	Status        string   `json:"status"`
	StorageClass  string   `json:"storageClass"`
	ClaimRef      string   `json:"claimRef"`
	Age           string   `json:"age"`
}

// PersistentVolumeClaimInfo persistentvolumeclaim 信息
type PersistentVolumeClaimInfo struct {
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace"`
	StorageClass string   `json:"storageClass"`
	Status       string   `json:"status"`
	Volume       string   `json:"volume"`
	AccessModes  []string `json:"accessModes"`
	Capacity     string   `json:"capacity"`
	Age          string   `json:"age"`
}

// StorageClassInfo storageclass 信息
type StorageClassInfo struct {
	Name              string `json:"name"`
	Provisioner       string `json:"provisioner"`
	ReclaimPolicy     string `json:"reclaimPolicy"`
	VolumeBindingMode string `json:"volumeBindingMode"`
	Age               string `json:"age"`
}

// ClusterStats 集群统计信息
type ClusterStats struct {
	Nodes       int            `json:"nodes"`
	Namespaces  int            `json:"namespaces"`
	Pods        int            `json:"pods"`
	Deployments int            `json:"deployments"`
	StatefulSets int           `json:"statefulSets"`
	DaemonSets  int            `json:"daemonSets"`
	PVs         int            `json:"pvs"`
	PVCs        int            `json:"pvcs"`
	Version     string         `json:"version"`
	NodeStats   []NodeStatInfo `json:"nodeStats"`
}

// NodeStatInfo 节点统计
type NodeStatInfo struct {
	Name            string `json:"name"`
	Status          string `json:"status"`
	Pods            int    `json:"pods"`
	PodCapacity     int    `json:"podCapacity"`
	CPUAllocatable  string `json:"cpuAllocatable"`
	MemoryAllocatable string `json:"memoryAllocatable"`
}

// GetClusterStats 获取集群统计信息
func (s *K8sService) GetClusterStats(ctx context.Context) (*ClusterStats, error) {
	stats := &ClusterStats{}

	// 获取 K8s 版本
	version, err := s.discoveryClient.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("获取集群版本失败: %w", err)
	}
	stats.Version = fmt.Sprintf("%s.%s", version.Major, version.Minor)

	// 并发获取各类资源数量，每个 goroutine 返回独立结果，避免数据竞争
	type countResult struct {
		key       string
		count     int
		nodeStats []NodeStatInfo
		err       error
	}

	results := make([]countResult, 8)

	var wg sync.WaitGroup
	wg.Add(8)

	// 节点
	go func() {
		defer wg.Done()
		list, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			results[0] = countResult{key: "nodes", err: err}
			return
		}
		nodeStats := make([]NodeStatInfo, 0, len(list.Items))
		for _, node := range list.Items {
			ns := NodeStatInfo{Name: node.Name}
			for _, cond := range node.Status.Conditions {
				if cond.Type == corev1.NodeReady {
					if cond.Status == corev1.ConditionTrue {
						ns.Status = "Ready"
					} else {
						ns.Status = "NotReady"
					}
					break
				}
			}
			if node.Status.Capacity != nil {
				if qty, ok := node.Status.Capacity[corev1.ResourcePods]; ok {
					ns.PodCapacity = int(qty.Value())
				}
				if qty, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
					ns.CPUAllocatable = qty.String()
				}
				if qty, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
					ns.MemoryAllocatable = qty.String()
				}
			}
			nodeStats = append(nodeStats, ns)
		}
		results[0] = countResult{key: "nodes", count: len(list.Items), nodeStats: nodeStats}
	}()

	// Namespace
	go func() {
		defer wg.Done()
		list, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			results[1] = countResult{key: "namespaces", err: err}
			return
		}
		results[1] = countResult{key: "namespaces", count: len(list.Items)}
	}()

	// Pods
	go func() {
		defer wg.Done()
		list, err := s.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			results[2] = countResult{key: "pods", err: err}
			return
		}
		results[2] = countResult{key: "pods", count: len(list.Items)}
	}()

	// Deployments
	go func() {
		defer wg.Done()
		list, err := s.client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
		if err != nil {
			results[3] = countResult{key: "deployments", err: err}
			return
		}
		results[3] = countResult{key: "deployments", count: len(list.Items)}
	}()

	// StatefulSets
	go func() {
		defer wg.Done()
		list, err := s.client.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
		if err != nil {
			results[4] = countResult{key: "statefulSets", err: err}
			return
		}
		results[4] = countResult{key: "statefulSets", count: len(list.Items)}
	}()

	// DaemonSets
	go func() {
		defer wg.Done()
		list, err := s.client.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
		if err != nil {
			results[5] = countResult{key: "daemonSets", err: err}
			return
		}
		results[5] = countResult{key: "daemonSets", count: len(list.Items)}
	}()

	// PVs
	go func() {
		defer wg.Done()
		list, err := s.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
		if err != nil {
			results[6] = countResult{key: "pvs", err: err}
			return
		}
		results[6] = countResult{key: "pvs", count: len(list.Items)}
	}()

	// PVCs
	go func() {
		defer wg.Done()
		list, err := s.client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			results[7] = countResult{key: "pvcs", err: err}
			return
		}
		results[7] = countResult{key: "pvcs", count: len(list.Items)}
	}()

	wg.Wait()

	// 在主 goroutine 中组装结果，无数据竞争
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		switch r.key {
		case "nodes":
			stats.Nodes = r.count
			stats.NodeStats = r.nodeStats
		case "namespaces":
			stats.Namespaces = r.count
		case "pods":
			stats.Pods = r.count
		case "deployments":
			stats.Deployments = r.count
		case "statefulSets":
			stats.StatefulSets = r.count
		case "daemonSets":
			stats.DaemonSets = r.count
		case "pvs":
			stats.PVs = r.count
		case "pvcs":
			stats.PVCs = r.count
		}
	}

	return stats, nil
}

// GetResourceYAML 获取资源的 YAML 定义
func (s *K8sService) GetResourceYAML(ctx context.Context, resourceType, namespace, name string) (string, error) {
	var obj interface{}
	var err error

	switch resourceType {
	case "pod", "pods":
		obj, err = s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	case "deployment", "deployments":
		obj, err = s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	case "statefulset", "statefulsets":
		obj, err = s.client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	case "daemonset", "daemonsets":
		obj, err = s.client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	case "configmap", "configmaps":
		obj, err = s.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	case "secret", "secrets":
		obj, err = s.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	case "persistentvolumeclaim", "persistentvolumeclaims", "pvc", "pvcs":
		obj, err = s.client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	case "persistentvolume", "persistentvolumes", "pv", "pvs":
		obj, err = s.client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	case "storageclass", "storageclasses", "sc":
		obj, err = s.client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	case "namespace", "namespaces":
		obj, err = s.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	case "service", "services":
		obj, err = s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	case "ingress", "ingresses":
		obj, err = s.client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	case "node", "nodes":
		obj, err = s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	default:
		return "", fmt.Errorf("不支持的资源类型: %s", resourceType)
	}

	if err != nil {
		return "", fmt.Errorf("获取资源 %s/%s 失败: %w", resourceType, name, err)
	}

	// 转换为 YAML
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return "", fmt.Errorf("获取 GVK 失败: %w", err)
	}
	if len(gvks) > 0 {
		obj.GetObjectKind().SetGroupVersionKind(gvks[0])
	}

	// 使用 k8s.io/apimachinery 的 serializer
	mediaType := "application/yaml"
	info, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return "", fmt.Errorf("不支持 YAML 序列化")
	}

	encoder := scheme.Codecs.EncoderForVersion(info.Serializer, schema.GroupVersion{Group: "", Version: "v1"})
	yamlBytes, err := runtime.Encode(encoder, obj)
	if err != nil {
		return "", fmt.Errorf("YAML 序列化失败: %w", err)
	}

	return string(yamlBytes), nil
}

// ListNamespaces 获取 namespace 列表
func (s *K8sService) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {
	list, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]NamespaceInfo, 0, len(list.Items))
	for _, ns := range list.Items {
		result = append(result, NamespaceInfo{
			Name:   ns.Name,
			Status: ns.Status.Phase,
			Age:    formatAge(ns.CreationTimestamp.Time),
			Labels: ns.Labels,
		})
	}
	return result, nil
}

// ListPods 获取 pod 列表，namespace 为空时查所有
func (s *K8sService) ListPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	list, err := s.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]PodInfo, 0, len(list.Items))
	for _, pod := range list.Items {
		result = append(result, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    getPodStatus(&pod),
			Restarts:  getTotalRestarts(pod.Status.ContainerStatuses),
			Node:      pod.Spec.NodeName,
			Age:       formatAge(pod.CreationTimestamp.Time),
			IP:        pod.Status.PodIP,
		})
	}
	return result, nil
}

// ListDeployments 获取 deployment 列表，namespace 为空时查所有
func (s *K8sService) ListDeployments(ctx context.Context, namespace string) ([]DeploymentInfo, error) {
	list, err := s.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]DeploymentInfo, 0, len(list.Items))
	for _, d := range list.Items {
		result = append(result, DeploymentInfo{
			Name:      d.Name,
			Namespace: d.Namespace,
			Replicas:  *d.Spec.Replicas,
			Ready:     fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, *d.Spec.Replicas),
			Age:       formatAge(d.CreationTimestamp.Time),
			Images:    getContainerImages(d.Spec.Template.Spec.Containers),
			Labels:    d.Labels,
		})
	}
	return result, nil
}

// ListStatefulSets 获取 statefulset 列表，namespace 为空时查所有
func (s *K8sService) ListStatefulSets(ctx context.Context, namespace string) ([]StatefulSetInfo, error) {
	list, err := s.client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]StatefulSetInfo, 0, len(list.Items))
	for _, ss := range list.Items {
		result = append(result, StatefulSetInfo{
			Name:      ss.Name,
			Namespace: ss.Namespace,
			Replicas:  *ss.Spec.Replicas,
			Ready:     fmt.Sprintf("%d/%d", ss.Status.ReadyReplicas, *ss.Spec.Replicas),
			Age:       formatAge(ss.CreationTimestamp.Time),
			Images:    getContainerImages(ss.Spec.Template.Spec.Containers),
			Labels:    ss.Labels,
		})
	}
	return result, nil
}

// ListDaemonSets 获取 daemonset 列表，namespace 为空时查所有
func (s *K8sService) ListDaemonSets(ctx context.Context, namespace string) ([]DaemonSetInfo, error) {
	list, err := s.client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]DaemonSetInfo, 0, len(list.Items))
	for _, ds := range list.Items {
		result = append(result, DaemonSetInfo{
			Name:      ds.Name,
			Namespace: ds.Namespace,
			Desired:   ds.Status.DesiredNumberScheduled,
			Ready:     ds.Status.NumberReady,
			Age:       formatAge(ds.CreationTimestamp.Time),
			Images:    getContainerImages(ds.Spec.Template.Spec.Containers),
			Labels:    ds.Labels,
		})
	}
	return result, nil
}

// ListConfigMaps 获取 configmap 列表，namespace 为空时查所有
func (s *K8sService) ListConfigMaps(ctx context.Context, namespace string) ([]ConfigMapInfo, error) {
	list, err := s.client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]ConfigMapInfo, 0, len(list.Items))
	for _, cm := range list.Items {
		keys := make([]string, 0, len(cm.Data))
		for k := range cm.Data {
			keys = append(keys, k)
		}
		result = append(result, ConfigMapInfo{
			Name:      cm.Name,
			Namespace: cm.Namespace,
			Keys:      keys,
			Age:       formatAge(cm.CreationTimestamp.Time),
			Labels:    cm.Labels,
		})
	}
	return result, nil
}

// ListSecrets 获取 secret 列表，namespace 为空时查所有
func (s *K8sService) ListSecrets(ctx context.Context, namespace string) ([]SecretInfo, error) {
	list, err := s.client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]SecretInfo, 0, len(list.Items))
	for _, sec := range list.Items {
		result = append(result, SecretInfo{
			Name:      sec.Name,
			Namespace: sec.Namespace,
			Type:      string(sec.Type),
			Age:       formatAge(sec.CreationTimestamp.Time),
			Labels:    sec.Labels,
		})
	}
	return result, nil
}

// ListPersistentVolumes 获取 persistentvolume 列表
func (s *K8sService) ListPersistentVolumes(ctx context.Context) ([]PersistentVolumeInfo, error) {
	list, err := s.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]PersistentVolumeInfo, 0, len(list.Items))
	for _, pv := range list.Items {
		capacity := ""
		if pv.Spec.Capacity != nil {
			if qty, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
				capacity = qty.String()
			}
		}

		accessModes := make([]string, 0, len(pv.Spec.AccessModes))
		for _, am := range pv.Spec.AccessModes {
			accessModes = append(accessModes, string(am))
		}

		claimRef := ""
		if pv.Spec.ClaimRef != nil {
			claimRef = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		}

		result = append(result, PersistentVolumeInfo{
			Name:          pv.Name,
			Capacity:      capacity,
			AccessModes:   accessModes,
			ReclaimPolicy: string(pv.Spec.PersistentVolumeReclaimPolicy),
			Status:        string(pv.Status.Phase),
			StorageClass:  pv.Spec.StorageClassName,
			ClaimRef:      claimRef,
			Age:           formatAge(pv.CreationTimestamp.Time),
		})
	}
	return result, nil
}

// ListPersistentVolumeClaims 获取 persistentvolumeclaim 列表，namespace 为空时查所有
func (s *K8sService) ListPersistentVolumeClaims(ctx context.Context, namespace string) ([]PersistentVolumeClaimInfo, error) {
	list, err := s.client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]PersistentVolumeClaimInfo, 0, len(list.Items))
	for _, pvc := range list.Items {
		accessModes := make([]string, 0, len(pvc.Spec.AccessModes))
		for _, am := range pvc.Spec.AccessModes {
			accessModes = append(accessModes, string(am))
		}

		capacity := ""
		if pvc.Status.Capacity != nil {
			if qty, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
				capacity = qty.String()
			}
		}

		volume := ""
		if pvc.Spec.VolumeName != "" {
			volume = pvc.Spec.VolumeName
		}

		storageClass := ""
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}

		result = append(result, PersistentVolumeClaimInfo{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			StorageClass: storageClass,
			Status:       string(pvc.Status.Phase),
			Volume:       volume,
			AccessModes:  accessModes,
			Capacity:     capacity,
			Age:          formatAge(pvc.CreationTimestamp.Time),
		})
	}
	return result, nil
}

// ListStorageClasses 获取 storageclass 列表
func (s *K8sService) ListStorageClasses(ctx context.Context) ([]StorageClassInfo, error) {
	list, err := s.client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]StorageClassInfo, 0, len(list.Items))
	for _, sc := range list.Items {
		reclaimPolicy := ""
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		bindingMode := ""
		if sc.VolumeBindingMode != nil {
			bindingMode = string(*sc.VolumeBindingMode)
		}

		result = append(result, StorageClassInfo{
			Name:              sc.Name,
			Provisioner:       sc.Provisioner,
			ReclaimPolicy:     reclaimPolicy,
			VolumeBindingMode: bindingMode,
			Age:               formatAge(sc.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func getContainerImages(containers []corev1.Container) []string {
	images := make([]string, 0, len(containers))
	for _, c := range containers {
		images = append(images, c.Image)
	}
	return images
}

func getPodStatus(pod *corev1.Pod) string {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
			return "Running"
		}
	}
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	switch pod.Status.Phase {
	case corev1.PodPending:
		return "Pending"
	case corev1.PodSucceeded:
		return "Succeeded"
	case corev1.PodFailed:
		return "Failed"
	case corev1.PodUnknown:
		return "Unknown"
	}
	return string(pod.Status.Phase)
}

func getTotalRestarts(statuses []corev1.ContainerStatus) int32 {
	var total int32
	for _, cs := range statuses {
		total += cs.RestartCount
	}
	return total
}

func formatAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	if d < time.Minute {
		return "0m"
	}
	if d < time.Hour {
		return formatDuration(d, time.Minute)
	}
	if d < 24*time.Hour {
		return formatDuration(d, time.Hour)
	}
	return formatDuration(d, 24*time.Hour) + "d"
}

func formatDuration(d, unit time.Duration) string {
	v := int(d / unit)
	return fmt.Sprintf("%d%c", v, unit.String()[0])
}
