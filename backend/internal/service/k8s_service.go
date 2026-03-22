package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	var obj runtime.Object
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

// PodDetailInfo Pod 详情信息
type PodDetailInfo struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Status          string            `json:"status"`
	Phase           string            `json:"phase"`
	Node            string            `json:"node"`
	NodeIP          string            `json:"nodeIP"`
	PodIP           string            `json:"podIP"`
	RestartCount    int32             `json:"restartCount"`
	Age             string            `json:"age"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	OwnerReferences []OwnerRef        `json:"ownerReferences"`
	Containers      []ContainerInfo   `json:"containers"`
	Conditions      []PodCondition    `json:"conditions"`
	Volumes         []VolumeInfo      `json:"volumes"`
}

// OwnerRef 所有者引用
type OwnerRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	Ready        bool              `json:"ready"`
	RestartCount int32             `json:"restartCount"`
	State        string            `json:"state"`
	Reason       string            `json:"reason"`
	Message      string            `json:"message"`
	Ports        []ContainerPort   `json:"ports"`
	Resources    ResourceInfo      `json:"resources"`
}

// ContainerPort 容器端口
type ContainerPort struct {
	Name          string `json:"name"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

// ResourceInfo 资源信息
type ResourceInfo struct {
	Limits   map[string]string `json:"limits"`
	Requests map[string]string `json:"requests"`
}

// PodCondition Pod 条件
type PodCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// VolumeInfo 卷信息
type VolumeInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

// GetPodDetail 获取 Pod 详情
func (s *K8sService) GetPodDetail(ctx context.Context, namespace, name string) (*PodDetailInfo, error) {
	pod, err := s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 获取 Node IP
	nodeIP := ""
	for _, ip := range pod.Status.HostIPs {
		nodeIP = ip.IP
		break
	}
	if nodeIP == "" && pod.Status.HostIP != "" {
		nodeIP = pod.Status.HostIP
	}

	// 构建容器信息
	containers := make([]ContainerInfo, 0, len(pod.Spec.Containers))
	for _, c := range pod.Spec.Containers {
		containerInfo := ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
			Resources: ResourceInfo{
				Limits:   make(map[string]string),
				Requests: make(map[string]string),
			},
		}

		// 资源限制
		for k, v := range c.Resources.Limits {
			containerInfo.Resources.Limits[string(k)] = v.String()
		}
		for k, v := range c.Resources.Requests {
			containerInfo.Resources.Requests[string(k)] = v.String()
		}

		// 端口
		for _, p := range c.Ports {
			containerInfo.Ports = append(containerInfo.Ports, ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				Protocol:      string(p.Protocol),
			})
		}

		// 查找容器状态
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == c.Name {
				containerInfo.Ready = cs.Ready
				containerInfo.RestartCount = cs.RestartCount
				containerInfo.State, containerInfo.Reason, containerInfo.Message = getContainerState(&cs)
				break
			}
		}

		containers = append(containers, containerInfo)
	}

	// 构建条件信息
	conditions := make([]PodCondition, 0, len(pod.Status.Conditions))
	for _, cond := range pod.Status.Conditions {
		conditions = append(conditions, PodCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	// 构建卷信息
	volumes := make([]VolumeInfo, 0, len(pod.Spec.Volumes))
	for _, vol := range pod.Spec.Volumes {
		volInfo := VolumeInfo{Name: vol.Name}
		if vol.ConfigMap != nil {
			volInfo.Type = "ConfigMap"
			volInfo.Source = vol.ConfigMap.Name
		} else if vol.Secret != nil {
			volInfo.Type = "Secret"
			volInfo.Source = vol.Secret.SecretName
		} else if vol.PersistentVolumeClaim != nil {
			volInfo.Type = "PVC"
			volInfo.Source = vol.PersistentVolumeClaim.ClaimName
		} else if vol.EmptyDir != nil {
			volInfo.Type = "EmptyDir"
		} else if vol.HostPath != nil {
			volInfo.Type = "HostPath"
			volInfo.Source = vol.HostPath.Path
		} else {
			volInfo.Type = "Other"
		}
		volumes = append(volumes, volInfo)
	}

	// 构建所有者引用
	ownerRefs := make([]OwnerRef, 0, len(pod.OwnerReferences))
	for _, ref := range pod.OwnerReferences {
		ownerRefs = append(ownerRefs, OwnerRef{
			APIVersion: ref.APIVersion,
			Kind:       ref.Kind,
			Name:       ref.Name,
		})
	}

	return &PodDetailInfo{
		Name:            pod.Name,
		Namespace:       pod.Namespace,
		Status:          getPodStatus(pod),
		Phase:           string(pod.Status.Phase),
		Node:            pod.Spec.NodeName,
		NodeIP:          nodeIP,
		PodIP:           pod.Status.PodIP,
		RestartCount:    getTotalRestarts(pod.Status.ContainerStatuses),
		Age:             formatAge(pod.CreationTimestamp.Time),
		Labels:          pod.Labels,
		Annotations:     pod.Annotations,
		OwnerReferences: ownerRefs,
		Containers:      containers,
		Conditions:      conditions,
		Volumes:         volumes,
	}, nil
}

func getContainerState(cs *corev1.ContainerStatus) (state, reason, message string) {
	if cs.State.Running != nil {
		return "Running", "", ""
	}
	if cs.State.Waiting != nil {
		return "Waiting", cs.State.Waiting.Reason, cs.State.Waiting.Message
	}
	if cs.State.Terminated != nil {
		return "Terminated", cs.State.Terminated.Reason, cs.State.Terminated.Message
	}
	return "Unknown", "", ""
}

// EventInfo 事件信息
type EventInfo struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Count     int32  `json:"count"`
	FirstTime string `json:"firstTime"`
	LastTime  string `json:"lastTime"`
	Source    string `json:"source"`
}

// GetPodEvents 获取 Pod 相关事件
func (s *K8sService) GetPodEvents(ctx context.Context, namespace, name string) ([]EventInfo, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", name, namespace)
	events, err := s.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	result := make([]EventInfo, 0, len(events.Items))
	for _, e := range events.Items {
		source := e.Source.Component
		if e.Source.Host != "" {
			source = fmt.Sprintf("%s, %s", source, e.Source.Host)
		}
		result = append(result, EventInfo{
			Type:      e.Type,
			Reason:    e.Reason,
			Message:   e.Message,
			Count:     e.Count,
			FirstTime: formatEventTime(e.FirstTimestamp.Time),
			LastTime:  formatEventTime(e.LastTimestamp.Time),
			Source:    source,
		})
	}
	return result, nil
}

func formatEventTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// GetPodLogs 获取 Pod 日志
func (s *K8sService) GetPodLogs(ctx context.Context, namespace, name, container string, tailLines int64, previous bool) (string, error) {
	options := &corev1.PodLogOptions{
		Container: container,
		Previous:  previous,
	}
	if tailLines > 0 {
		options.TailLines = &tailLines
	}

	req := s.client.CoreV1().Pods(namespace).GetLogs(name, options)
	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", err
	}
	return string(logs), nil
}

// DeletePod 删除 Pod
func (s *K8sService) DeletePod(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// DeploymentDetailInfo Deployment 详情信息
type DeploymentDetailInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	UpdatedReplicas   int32             `json:"updatedReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	Strategy          string            `json:"strategy"`
	Age               string            `json:"age"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Selector          map[string]string `json:"selector"`
	PodTemplate       PodTemplateInfo   `json:"podTemplate"`
	Conditions        []DeployCondition `json:"conditions"`
}

// PodTemplateInfo Pod 模板信息
type PodTemplateInfo struct {
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	Containers   []ContainerInfo   `json:"containers"`
	NodeSelector map[string]string `json:"nodeSelector"`
}

// DeployCondition Deployment 条件
type DeployCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// GetDeploymentDetail 获取 Deployment 详情
func (s *K8sService) GetDeploymentDetail(ctx context.Context, namespace, name string) (*DeploymentDetailInfo, error) {
	deploy, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 构建容器信息
	containers := make([]ContainerInfo, 0, len(deploy.Spec.Template.Spec.Containers))
	for _, c := range deploy.Spec.Template.Spec.Containers {
		containerInfo := ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
			Resources: ResourceInfo{
				Limits:   make(map[string]string),
				Requests: make(map[string]string),
			},
		}

		for k, v := range c.Resources.Limits {
			containerInfo.Resources.Limits[string(k)] = v.String()
		}
		for k, v := range c.Resources.Requests {
			containerInfo.Resources.Requests[string(k)] = v.String()
		}

		for _, p := range c.Ports {
			containerInfo.Ports = append(containerInfo.Ports, ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				Protocol:      string(p.Protocol),
			})
		}

		containers = append(containers, containerInfo)
	}

	// 构建条件信息
	conditions := make([]DeployCondition, 0, len(deploy.Status.Conditions))
	for _, cond := range deploy.Status.Conditions {
		conditions = append(conditions, DeployCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	strategy := "RollingUpdate"
	if deploy.Spec.Strategy.Type == appsv1.RecreateDeploymentStrategyType {
		strategy = "Recreate"
	}

	return &DeploymentDetailInfo{
		Name:              deploy.Name,
		Namespace:         deploy.Namespace,
		Replicas:          *deploy.Spec.Replicas,
		ReadyReplicas:     deploy.Status.ReadyReplicas,
		UpdatedReplicas:   deploy.Status.UpdatedReplicas,
		AvailableReplicas: deploy.Status.AvailableReplicas,
		Strategy:          strategy,
		Age:               formatAge(deploy.CreationTimestamp.Time),
		Labels:            deploy.Labels,
		Annotations:       deploy.Annotations,
		Selector:          deploy.Spec.Selector.MatchLabels,
		PodTemplate: PodTemplateInfo{
			Labels:       deploy.Spec.Template.Labels,
			Annotations:  deploy.Spec.Template.Annotations,
			Containers:   containers,
			NodeSelector: deploy.Spec.Template.Spec.NodeSelector,
		},
		Conditions: conditions,
	}, nil
}

// ScaleDeployment 扩缩容 Deployment
func (s *K8sService) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replicas,
		},
	}
	_, err := s.client.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	return err
}

// RestartDeployment 重启 Deployment（通过 annotation 触发滚动更新）
func (s *K8sService) RestartDeployment(ctx context.Context, namespace, name string) error {
	deploy, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = s.client.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	return err
}

// DeleteDeployment 删除 Deployment
func (s *K8sService) DeleteDeployment(ctx context.Context, namespace, name string) error {
	return s.client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ServiceInfo Service 信息
type ServiceInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"clusterIP"`
	ExternalIP  []string          `json:"externalIP"`
	Ports       []ServicePortInfo `json:"ports"`
	Selector    map[string]string `json:"selector"`
	Age         string            `json:"age"`
	Labels      map[string]string `json:"labels"`
}

// ServicePortInfo Service 端口信息
type ServicePortInfo struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	NodePort   int32  `json:"nodePort"`
	Protocol   string `json:"protocol"`
}

// ListServices 获取 Service 列表
func (s *K8sService) ListServices(ctx context.Context, namespace string) ([]ServiceInfo, error) {
	services, err := s.client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]ServiceInfo, 0, len(services.Items))
	for _, svc := range services.Items {
		externalIPs := make([]string, 0)
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, ip := range svc.Status.LoadBalancer.Ingress {
				if ip.IP != "" {
					externalIPs = append(externalIPs, ip.IP)
				}
				if ip.Hostname != "" {
					externalIPs = append(externalIPs, ip.Hostname)
				}
			}
		} else {
			externalIPs = svc.Spec.ExternalIPs
		}

		ports := make([]ServicePortInfo, 0, len(svc.Spec.Ports))
		for _, p := range svc.Spec.Ports {
			ports = append(ports, ServicePortInfo{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: p.TargetPort.String(),
				NodePort:   p.NodePort,
				Protocol:   string(p.Protocol),
			})
		}

		result = append(result, ServiceInfo{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: externalIPs,
			Ports:      ports,
			Selector:   svc.Spec.Selector,
			Age:        formatAge(svc.CreationTimestamp.Time),
			Labels:     svc.Labels,
		})
	}
	return result, nil
}

// GetServiceDetail 获取 Service 详情
func (s *K8sService) GetServiceDetail(ctx context.Context, namespace, name string) (*ServiceInfo, error) {
	svc, err := s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	externalIPs := make([]string, 0)
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ip := range svc.Status.LoadBalancer.Ingress {
			if ip.IP != "" {
				externalIPs = append(externalIPs, ip.IP)
			}
			if ip.Hostname != "" {
				externalIPs = append(externalIPs, ip.Hostname)
			}
		}
	} else {
		externalIPs = svc.Spec.ExternalIPs
	}

	ports := make([]ServicePortInfo, 0, len(svc.Spec.Ports))
	for _, p := range svc.Spec.Ports {
		ports = append(ports, ServicePortInfo{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: p.TargetPort.String(),
			NodePort:   p.NodePort,
			Protocol:   string(p.Protocol),
		})
	}

	return &ServiceInfo{
		Name:       svc.Name,
		Namespace:  svc.Namespace,
		Type:       string(svc.Spec.Type),
		ClusterIP:  svc.Spec.ClusterIP,
		ExternalIP: externalIPs,
		Ports:      ports,
		Selector:   svc.Spec.Selector,
		Age:        formatAge(svc.CreationTimestamp.Time),
		Labels:     svc.Labels,
	}, nil
}

// DeleteService 删除 Service
func (s *K8sService) DeleteService(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// IngressInfo Ingress 信息
type IngressInfo struct {
	Name      string         `json:"name"`
	Namespace string         `json:"namespace"`
	Class     string         `json:"class"`
	Hosts     []string       `json:"hosts"`
	Addresses []string       `json:"addresses"`
	Rules     []IngressRule  `json:"rules"`
	Age       string         `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// IngressRule Ingress 规则
type IngressRule struct {
	Host    string         `json:"host"`
	Paths   []IngressPath  `json:"paths"`
}

// IngressPath Ingress 路径
type IngressPath struct {
	Path        string `json:"path"`
	PathType    string `json:"pathType"`
	ServiceName string `json:"serviceName"`
	ServicePort string `json:"servicePort"`
}

// ListIngresses 获取 Ingress 列表
func (s *K8sService) ListIngresses(ctx context.Context, namespace string) ([]IngressInfo, error) {
	ingresses, err := s.client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]IngressInfo, 0, len(ingresses.Items))
	for _, ing := range ingresses.Items {
		hosts := make([]string, 0)
		rules := make([]IngressRule, 0, len(ing.Spec.Rules))
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
			paths := make([]IngressPath, 0, len(rule.HTTP.Paths))
			for _, p := range rule.HTTP.Paths {
				svcName := ""
				svcPort := ""
				if p.Backend.Service != nil {
					svcName = p.Backend.Service.Name
					if p.Backend.Service.Port.Number != 0 {
						svcPort = fmt.Sprintf("%d", p.Backend.Service.Port.Number)
					} else {
						svcPort = p.Backend.Service.Port.Name
					}
				}
				pathType := string(*p.PathType)
				if pathType == "" {
					pathType = "ImplementationSpecific"
				}
				paths = append(paths, IngressPath{
					Path:        p.Path,
					PathType:    pathType,
					ServiceName: svcName,
					ServicePort: svcPort,
				})
			}
			rules = append(rules, IngressRule{
				Host:  rule.Host,
				Paths: paths,
			})
		}

		addresses := make([]string, 0)
		for _, addr := range ing.Status.LoadBalancer.Ingress {
			if addr.IP != "" {
				addresses = append(addresses, addr.IP)
			}
			if addr.Hostname != "" {
				addresses = append(addresses, addr.Hostname)
			}
		}

		class := ""
		if ing.Spec.IngressClassName != nil {
			class = *ing.Spec.IngressClassName
		}

		result = append(result, IngressInfo{
			Name:      ing.Name,
			Namespace: ing.Namespace,
			Class:     class,
			Hosts:     hosts,
			Addresses: addresses,
			Rules:     rules,
			Age:       formatAge(ing.CreationTimestamp.Time),
			Labels:    ing.Labels,
		})
	}
	return result, nil
}

// GetIngressDetail 获取 Ingress 详情
func (s *K8sService) GetIngressDetail(ctx context.Context, namespace, name string) (*IngressInfo, error) {
	ing, err := s.client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	hosts := make([]string, 0)
	rules := make([]IngressRule, 0, len(ing.Spec.Rules))
	for _, rule := range ing.Spec.Rules {
		if rule.Host != "" {
			hosts = append(hosts, rule.Host)
		}
		paths := make([]IngressPath, 0, len(rule.HTTP.Paths))
		for _, p := range rule.HTTP.Paths {
			svcName := ""
			svcPort := ""
			if p.Backend.Service != nil {
				svcName = p.Backend.Service.Name
				if p.Backend.Service.Port.Number != 0 {
					svcPort = fmt.Sprintf("%d", p.Backend.Service.Port.Number)
				} else {
					svcPort = p.Backend.Service.Port.Name
				}
			}
			pathType := string(*p.PathType)
			if pathType == "" {
				pathType = "ImplementationSpecific"
			}
			paths = append(paths, IngressPath{
				Path:        p.Path,
				PathType:    pathType,
				ServiceName: svcName,
				ServicePort: svcPort,
			})
		}
		rules = append(rules, IngressRule{
			Host:  rule.Host,
			Paths: paths,
		})
	}

	addresses := make([]string, 0)
	for _, addr := range ing.Status.LoadBalancer.Ingress {
		if addr.IP != "" {
			addresses = append(addresses, addr.IP)
		}
		if addr.Hostname != "" {
			addresses = append(addresses, addr.Hostname)
		}
	}

	class := ""
	if ing.Spec.IngressClassName != nil {
		class = *ing.Spec.IngressClassName
	}

	return &IngressInfo{
		Name:      ing.Name,
		Namespace: ing.Namespace,
		Class:     class,
		Hosts:     hosts,
		Addresses: addresses,
		Rules:     rules,
		Age:       formatAge(ing.CreationTimestamp.Time),
		Labels:    ing.Labels,
	}, nil
}

// DeleteIngress 删除 Ingress
func (s *K8sService) DeleteIngress(ctx context.Context, namespace, name string) error {
	return s.client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// NodeInfo Node 信息
type NodeInfo struct {
	Name           string            `json:"name"`
	Status         string            `json:"status"`
	Roles          []string          `json:"roles"`
	Version        string            `json:"version"`
	OSImage        string            `json:"osImage"`
	KernelVersion  string            `json:"kernelVersion"`
	ContainerRuntime string          `json:"containerRuntime"`
	CPU            string            `json:"cpu"`
	Memory         string            `json:"memory"`
	Age            string            `json:"age"`
	Labels         map[string]string `json:"labels"`
}

// ListNodes 获取 Node 列表
func (s *K8sService) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]NodeInfo, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		// 获取状态
		status := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				if cond.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		// 获取角色
		roles := make([]string, 0)
		for k := range node.Labels {
			if strings.HasPrefix(k, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
				if role != "" {
					roles = append(roles, role)
				}
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "<none>")
		}

		// 获取资源
		cpu := node.Status.Capacity.Cpu().String()
		memory := node.Status.Capacity.Memory().String()

		result = append(result, NodeInfo{
			Name:             node.Name,
			Status:           status,
			Roles:            roles,
			Version:          node.Status.NodeInfo.KubeletVersion,
			OSImage:          node.Status.NodeInfo.OSImage,
			KernelVersion:    node.Status.NodeInfo.KernelVersion,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			CPU:              cpu,
			Memory:           memory,
			Age:              formatAge(node.CreationTimestamp.Time),
			Labels:           node.Labels,
		})
	}
	return result, nil
}

// NodeDetailInfo Node 详情信息
type NodeDetailInfo struct {
	Name               string                 `json:"name"`
	Status             string                 `json:"status"`
	Roles              []string               `json:"roles"`
	Version            string                 `json:"version"`
	OSImage            string                 `json:"osImage"`
	KernelVersion      string                 `json:"kernelVersion"`
	ContainerRuntime   string                 `json:"containerRuntime"`
	Architecture       string                 `json:"architecture"`
	OperatingSystem    string                 `json:"operatingSystem"`
	CPU                string                 `json:"cpu"`
	Memory             string                 `json:"memory"`
	EphemeralStorage   string                 `json:"ephemeralStorage"`
	Pods               string                 `json:"pods"`
	Age                string                 `json:"age"`
	Labels             map[string]string      `json:"labels"`
	Annotations        map[string]string      `json:"annotations"`
	Addresses          []NodeAddress          `json:"addresses"`
	Conditions         []NodeCondition        `json:"conditions"`
	AllocatedResources AllocatedResourceInfo  `json:"allocatedResources"`
}

// NodeAddress Node 地址
type NodeAddress struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

// NodeCondition Node 条件
type NodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// AllocatedResourceInfo 已分配资源信息
type AllocatedResourceInfo struct {
	CPU              string `json:"cpu"`
	Memory           string `json:"memory"`
	EphemeralStorage string `json:"ephemeralStorage"`
	Pods             int    `json:"pods"`
}

// GetNodeDetail 获取 Node 详情
func (s *K8sService) GetNodeDetail(ctx context.Context, name string) (*NodeDetailInfo, error) {
	node, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 获取状态
	status := "Unknown"
	conditions := make([]NodeCondition, 0, len(node.Status.Conditions))
	for _, cond := range node.Status.Conditions {
		conditions = append(conditions, NodeCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
		}
	}

	// 获取角色
	roles := make([]string, 0)
	for k := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
			if role != "" {
				roles = append(roles, role)
			}
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "<none>")
	}

	// 获取地址
	addresses := make([]NodeAddress, 0, len(node.Status.Addresses))
	for _, addr := range node.Status.Addresses {
		addresses = append(addresses, NodeAddress{
			Type:    string(addr.Type),
			Address: addr.Address,
		})
	}

	// 计算已分配资源
	pods, err := s.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", name),
	})
	allocatedCPU := resource.Quantity{}
	allocatedMemory := resource.Quantity{}
	allocatedStorage := resource.Quantity{}
	allocatedPods := 0
	if err == nil {
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				continue
			}
			allocatedPods++
			for _, c := range pod.Spec.Containers {
				if c.Resources.Requests != nil {
					if cpu, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
						allocatedCPU.Add(cpu)
					}
					if mem, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
						allocatedMemory.Add(mem)
					}
					if storage, ok := c.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
						allocatedStorage.Add(storage)
					}
				}
			}
		}
	}

	return &NodeDetailInfo{
		Name:             node.Name,
		Status:           status,
		Roles:            roles,
		Version:          node.Status.NodeInfo.KubeletVersion,
		OSImage:          node.Status.NodeInfo.OSImage,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		Architecture:     node.Status.NodeInfo.Architecture,
		OperatingSystem:  node.Status.NodeInfo.OperatingSystem,
		CPU:              node.Status.Capacity.Cpu().String(),
		Memory:           node.Status.Capacity.Memory().String(),
		EphemeralStorage: node.Status.Capacity.StorageEphemeral().String(),
		Pods:             node.Status.Capacity.Pods().String(),
		Age:              formatAge(node.CreationTimestamp.Time),
		Labels:           node.Labels,
		Annotations:      node.Annotations,
		Addresses:        addresses,
		Conditions:       conditions,
		AllocatedResources: AllocatedResourceInfo{
			CPU:              allocatedCPU.String(),
			Memory:           allocatedMemory.String(),
			EphemeralStorage: allocatedStorage.String(),
			Pods:             allocatedPods,
		},
	}, nil
}
