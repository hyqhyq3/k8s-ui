package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8sService Kubernetes 业务逻辑
type K8sService struct {
	client *kubernetes.Clientset
}

// NewK8sService 创建 K8s 服务
func NewK8sService(client *kubernetes.Clientset) *K8sService {
	return &K8sService{client: client}
}

// NamespaceInfo namespace 信息
type NamespaceInfo struct {
	Name   string            `json:"name"`
	Status corev1.NamespacePhase `json:"status"`
	Age    string            `json:"age"`
	Labels map[string]string `json:"labels"`
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
