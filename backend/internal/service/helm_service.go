package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yangqihuang/k8s-ui/internal/helm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// HelmService Helm 业务逻辑（CRD 驱动，通过 Flux Helm Controller 执行实际操作）
type HelmService struct {
	client *helm.CRDClient
}

// NewHelmService 创建 Helm 服务
func NewHelmService(client *helm.CRDClient) *HelmService {
	return &HelmService{client: client}
}

// InstallOptions 安装选项
type InstallOptions struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Chart     string `json:"chart"`
	Repo      string `json:"repo"`
	Version   string `json:"version"`
	Values    string `json:"values"`
	Wait      bool   `json:"wait"`
}

// UpgradeOptions 升级选项
type UpgradeOptions struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Chart       string `json:"chart"`
	Repo        string `json:"repo"`
	Version     string `json:"version"`
	Values      string `json:"values"`
	Wait        bool   `json:"wait"`
	ResetValues bool   `json:"resetValues"`
	ReuseValues bool   `json:"reuseValues"`
}

// HelmReleaseInfo release 信息（列表用）
type HelmReleaseInfo struct {
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	Status      string    `json:"status"`
	Chart       string    `json:"chart"`
	AppVersion  string    `json:"appVersion"`
	Revision    int       `json:"revision"`
	Updated     time.Time `json:"updated"`
	Description string    `json:"description"`
}

// HelmReleaseDetail release 详情
type HelmReleaseDetail struct {
	HelmReleaseInfo
	Values       string `json:"values"`
	ChartName    string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
	Notes        string `json:"notes"`
}

// HelmReleaseHistory revision 历史条目
type HelmReleaseHistory struct {
	Revision    int       `json:"revision"`
	Chart       string    `json:"chart"`
	AppVersion  string    `json:"appVersion"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Updated     time.Time `json:"updated"`
}

// HelmRepoInfo chart repo 信息
type HelmRepoInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// DriftResource represents a single Kubernetes object drift
type DriftResource struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace"`
	APIVersion string `json:"apiVersion"`
	Drifted    bool   `json:"drifted"`
}

// DriftResult contains drift analysis for a Helm release
type DriftResult struct {
	ReleaseName string          `json:"releaseName"`
	Namespace   string          `json:"namespace"`
	Drifted     bool            `json:"drifted"`
	Resources   []DriftResource `json:"resources"`
	Summary     string          `json:"summary"`
}

// CheckDrift 通过 HelmRelease CRD 的 status.conditions 检测漂移
// Flux Helm Controller 在 driftDetection.mode=enabled 时会自动设置 Drifted=True 条件
func (s *HelmService) CheckDrift(ctx context.Context, namespace, name string) (*DriftResult, error) {
	hr, err := s.client.GetRelease(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取 HelmRelease 失败: %w", err)
	}

	conditions, _, _ := unstructured.NestedSlice(hr.Object, "status", "conditions")
	drifted := false
	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _, _ := unstructured.NestedString(cond, "type")
		status, _, _ := unstructured.NestedString(cond, "status")
		if condType == "Drifted" && status == "True" {
			drifted = true
			break
		}
	}

	// 从 status.inventory.entries 读取资源清单
	entries, _, _ := unstructured.NestedSlice(hr.Object, "status", "inventory", "entries")
	driftResources := make([]DriftResource, 0, len(entries))
	for _, e := range entries {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		id, _, _ := unstructured.NestedString(entry, "id")
		driftResources = append(driftResources, DriftResource{
			Name:      id,
			Namespace: namespace,
		})
	}

	summary := "无漂移"
	if drifted {
		summary = "检测到漂移，Flux Helm Controller 将自动修正"
	}

	return &DriftResult{
		ReleaseName: name,
		Namespace:   namespace,
		Drifted:     drifted,
		Resources:   driftResources,
		Summary:     summary,
	}, nil
}

// ListReleases 列出 HelmRelease CRD
func (s *HelmService) ListReleases(ctx context.Context, namespace string) ([]HelmReleaseInfo, error) {
	list, err := s.client.ListReleases(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取 release 列表失败: %w", err)
	}

	result := make([]HelmReleaseInfo, 0, len(list.Items))
	for _, item := range list.Items {
		info := parseHelmReleaseInfo(&item)
		result = append(result, info)
	}
	return result, nil
}

// GetRelease 获取 HelmRelease CRD 详情
func (s *HelmService) GetRelease(ctx context.Context, namespace, name string) (*HelmReleaseDetail, error) {
	hr, err := s.client.GetRelease(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取 release 详情失败: %w", err)
	}

	return parseHelmReleaseDetail(hr), nil
}

// GetReleaseHistory 获取 release revision 历史（从 status.history 读取）
func (s *HelmService) GetReleaseHistory(ctx context.Context, namespace, name string) ([]HelmReleaseHistory, error) {
	hr, err := s.client.GetRelease(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取 release 历史失败: %w", err)
	}

	historySlice, _, _ := unstructured.NestedSlice(hr.Object, "status", "history")
	result := make([]HelmReleaseHistory, 0, len(historySlice))
	for _, h := range historySlice {
		entry, ok := h.(map[string]interface{})
		if !ok {
			continue
		}

		revision := nestedInt(entry, "version")
		chartName, _, _ := unstructured.NestedString(entry, "chartName")
		chartVersion, _, _ := unstructured.NestedString(entry, "chartVersion")
		appVersion, _, _ := unstructured.NestedString(entry, "appVersion")
		status, _, _ := unstructured.NestedString(entry, "status")
		description, _, _ := unstructured.NestedString(entry, "description")
		updatedStr, _, _ := unstructured.NestedString(entry, "firstDeployed")

		updated := parseTime(updatedStr)

		result = append(result, HelmReleaseHistory{
			Revision:    revision,
			Chart:       fmt.Sprintf("%s-%s", chartName, chartVersion),
			AppVersion:  appVersion,
			Status:      formatCRDStatus(status),
			Description: description,
			Updated:     updated,
		})
	}
	return result, nil
}

// UninstallRelease 删除 HelmRelease CRD（Flux controller 会自动执行 helm uninstall）
func (s *HelmService) UninstallRelease(ctx context.Context, namespace, name string, keepHistory bool) error {
	err := s.client.DeleteRelease(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("卸载 release 失败: %w", err)
	}
	return nil
}

// RollbackRelease 通过 patch chart version 触发回滚
func (s *HelmService) RollbackRelease(ctx context.Context, namespace, name string, revision int) error {
	err := s.client.RollbackRelease(ctx, namespace, name, revision)
	if err != nil {
		return fmt.Errorf("回滚 release 失败: %w", err)
	}
	return nil
}

// InstallRelease 创建 HelmRelease CRD（Flux controller 会自动执行 helm install）
func (s *HelmService) InstallRelease(ctx context.Context, opts *InstallOptions) (*HelmReleaseDetail, error) {
	// 解析 values
	var valuesMap map[string]interface{}
	if opts.Values != "" {
		if err := yaml.Unmarshal([]byte(opts.Values), &valuesMap); err != nil {
			return nil, fmt.Errorf("解析 values YAML 失败: %w", err)
		}
	}
	if valuesMap == nil {
		valuesMap = make(map[string]interface{})
	}

	// 构建 HelmRelease CRD
	hr := &unstructured.Unstructured{}
	hr.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "helm.toolkit.fluxcd.io",
		Version: "v2",
		Kind:    "HelmRelease",
	})
	hr.SetName(opts.Name)
	hr.SetNamespace(opts.Namespace)
	hr.Object["spec"] = map[string]interface{}{
		"interval": "5m",
		"chart": map[string]interface{}{
			"spec": map[string]interface{}{
				"chart": opts.Chart,
			},
		},
		"values": valuesMap,
		"install": map[string]interface{}{
			"remediation": map[string]interface{}{
				"retries": 3,
			},
		},
		"upgrade": map[string]interface{}{
			"remediation": map[string]interface{}{
				"retries": 3,
			},
		},
		"driftDetection": map[string]interface{}{
			"mode": "enabled",
		},
	}

	// 设置版本（如果有）
	if opts.Version != "" {
		chartSpec, _, _ := unstructured.NestedMap(hr.Object, "spec", "chart", "spec")
		chartSpec["version"] = opts.Version
	}

	// 设置 sourceRef（如果有 repo）
	if opts.Repo != "" {
		chartSpec, _, _ := unstructured.NestedMap(hr.Object, "spec", "chart", "spec")
		chartSpec["sourceRef"] = map[string]interface{}{
			"kind":      "HelmRepository",
			"name":      opts.Repo,
			"namespace": "flux-system",
		}
	}

	created, err := s.client.CreateRelease(ctx, opts.Namespace, hr)
	if err != nil {
		return nil, fmt.Errorf("安装 release 失败: %w", err)
	}

	return parseHelmReleaseDetail(created), nil
}

// UpgradeRelease 通过更新 HelmRelease CRD 触发升级
func (s *HelmService) UpgradeRelease(ctx context.Context, opts *UpgradeOptions) (*HelmReleaseDetail, error) {
	hr, err := s.client.GetRelease(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return nil, fmt.Errorf("获取 release 失败: %w", err)
	}

	// 解析新 values
	var valuesMap map[string]interface{}
	if opts.Values != "" {
		if err := yaml.Unmarshal([]byte(opts.Values), &valuesMap); err != nil {
			return nil, fmt.Errorf("解析 values YAML 失败: %w", err)
		}
	}
	if valuesMap == nil {
		valuesMap = make(map[string]interface{})
	}

	// 更新 spec
	spec, _, _ := unstructured.NestedMap(hr.Object, "spec")
	spec["values"] = valuesMap

	// 更新版本（如果有）
	if opts.Version != "" {
		chartSpec, ok := spec["chart"].(map[string]interface{})
		if ok {
			chartSpecInner, ok := chartSpec["spec"].(map[string]interface{})
			if ok {
				chartSpecInner["version"] = opts.Version
			}
		}
	}

	// 更新 sourceRef（如果有 repo）
	if opts.Repo != "" {
		chartSpec, ok := spec["chart"].(map[string]interface{})
		if ok {
			chartSpecInner, ok := chartSpec["spec"].(map[string]interface{})
			if ok {
				chartSpecInner["sourceRef"] = map[string]interface{}{
					"kind":      "HelmRepository",
					"name":      opts.Repo,
					"namespace": "flux-system",
				}
			}
		}
	}

	updated, err := s.client.UpdateRelease(ctx, opts.Namespace, hr)
	if err != nil {
		return nil, fmt.Errorf("升级 release 失败: %w", err)
	}

	return parseHelmReleaseDetail(updated), nil
}

// ListRepos 列出 HelmRepository CRD
func (s *HelmService) ListRepos(ctx context.Context) ([]HelmRepoInfo, error) {
	repos, err := s.client.ListRepos(ctx)
	if err != nil {
		return nil, fmt.Errorf("加载 repo 列表失败: %w", err)
	}

	result := make([]HelmRepoInfo, 0, len(repos.Items))
	for _, item := range repos.Items {
		url, _, _ := unstructured.NestedString(item.Object, "spec", "url")
		result = append(result, HelmRepoInfo{
			Name: item.GetName(),
			URL:  url,
		})
	}
	return result, nil
}

// AddRepo 创建 HelmRepository CRD
func (s *HelmService) AddRepo(ctx context.Context, name, url string) error {
	// 检查是否已存在
	_, err := s.client.GetRepo(ctx, name)
	if err == nil {
		return fmt.Errorf("repo %s 已存在", name)
	}

	repoObj := &unstructured.Unstructured{}
	repoObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "source.toolkit.fluxcd.io",
		Version: "v1",
		Kind:    "HelmRepository",
	})
	repoObj.SetName(name)
	repoObj.SetNamespace("flux-system")
	repoObj.Object["spec"] = map[string]interface{}{
		"url":      url,
		"interval": "5m",
	}

	_, err = s.client.CreateRepo(ctx, repoObj)
	if err != nil {
		return fmt.Errorf("创建 repo 失败: %w", err)
	}
	return nil
}

// RemoveRepo 删除 HelmRepository CRD
func (s *HelmService) RemoveRepo(ctx context.Context, name string) error {
	// 检查是否存在
	_, err := s.client.GetRepo(ctx, name)
	if err != nil {
		return fmt.Errorf("repo %s 不存在", name)
	}

	return s.client.DeleteRepo(ctx, name)
}

// SearchChart 通过 HTTP 下载 repo index.yaml 搜索 chart
func (s *HelmService) SearchChart(ctx context.Context, repoName, keyword string) ([]map[string]interface{}, error) {
	repo, err := s.client.GetRepo(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("获取 repo %s 失败: %w", repoName, err)
	}

	url, _, _ := unstructured.NestedString(repo.Object, "spec", "url")
	if url == "" {
		return nil, fmt.Errorf("repo %s 的 URL 为空", repoName)
	}

	// 下载 index.yaml
	indexURL := strings.TrimSuffix(url, "/") + "/index.yaml"
	indexData, err := fetchURL(ctx, indexURL)
	if err != nil {
		return nil, fmt.Errorf("下载 repo index 失败: %w", err)
	}

	// 解析 index.yaml
	var index struct {
		Entries map[string][]struct {
			Version     string   `yaml:"version"`
			AppVersion  string   `yaml:"appVersion"`
			Description string   `yaml:"description"`
			Keywords    []string `yaml:"keywords"`
		} `yaml:"entries"`
	}
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		return nil, fmt.Errorf("解析 repo index 失败: %w", err)
	}

	results := make([]map[string]interface{}, 0)
	for name, versions := range index.Entries {
		if len(versions) == 0 {
			continue
		}
		latest := versions[0]

		matched := false
		if keyword != "" {
			lowerName := strings.ToLower(name)
			lowerDesc := strings.ToLower(latest.Description)
			lowerKeyword := strings.ToLower(keyword)
			if strings.Contains(lowerName, lowerKeyword) || strings.Contains(lowerDesc, lowerKeyword) {
				matched = true
			}
		} else {
			matched = true
		}

		if matched {
			results = append(results, map[string]interface{}{
				"name":        name,
				"version":     latest.Version,
				"appVersion":  latest.AppVersion,
				"description": latest.Description,
				"repo":        repoName,
			})
		}
	}
	return results, nil
}

// GetChartVersions 获取 chart 的可用版本列表
func (s *HelmService) GetChartVersions(ctx context.Context, repoName, chartName string) ([]map[string]string, error) {
	repo, err := s.client.GetRepo(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("获取 repo %s 失败: %w", repoName, err)
	}

	url, _, _ := unstructured.NestedString(repo.Object, "spec", "url")
	if url == "" {
		return nil, fmt.Errorf("repo %s 的 URL 为空", repoName)
	}

	indexURL := strings.TrimSuffix(url, "/") + "/index.yaml"
	indexData, err := fetchURL(ctx, indexURL)
	if err != nil {
		return nil, fmt.Errorf("下载 repo index 失败: %w", err)
	}

	var index struct {
		Entries map[string][]struct {
			Version    string `yaml:"version"`
			AppVersion string `yaml:"appVersion"`
			Created    string `yaml:"created"`
		} `yaml:"entries"`
	}
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		return nil, fmt.Errorf("解析 repo index 失败: %w", err)
	}

	chartVersions, exists := index.Entries[chartName]
	if !exists {
		return nil, fmt.Errorf("chart %s 在 repo %s 中不存在", chartName, repoName)
	}

	results := make([]map[string]string, 0, len(chartVersions))
	for _, v := range chartVersions {
		results = append(results, map[string]string{
			"version":    v.Version,
			"appVersion": v.AppVersion,
			"created":    v.Created,
		})
	}
	return results, nil
}

// --- helpers ---

func parseHelmReleaseInfo(item *unstructured.Unstructured) HelmReleaseInfo {
	// 从 status.conditions 提取 Ready 状态
	status := "unknown"
	conditions, _, _ := unstructured.NestedSlice(item.Object, "status", "conditions")
	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _, _ := unstructured.NestedString(cond, "type")
		condStatus, _, _ := unstructured.NestedString(cond, "status")
		if condType == "Ready" {
			if condStatus == "True" {
				status = "deployed"
			} else {
				status = "failed"
			}
			break
		}
	}

	// 从 spec.chart.spec 提取 chart 信息
	chartName, _, _ := unstructured.NestedString(item.Object, "spec", "chart", "spec", "chart")

	// 从 status 提取 revision
	revision := nestedInt(item.Object, "status", "lastReleaseRevision")

	// 从 status.conditions 提取最后更新时间
	var updated time.Time
	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _, _ := unstructured.NestedString(cond, "type")
		if condType == "Ready" {
			ts, _, _ := unstructured.NestedString(cond, "lastTransitionTime")
			updated = parseTime(ts)
			break
		}
	}

	return HelmReleaseInfo{
		Name:        item.GetName(),
		Namespace:   item.GetNamespace(),
		Status:      status,
		Chart:       chartName,
		Revision:    revision,
		Updated:     updated,
		Description: "",
	}
}

func parseHelmReleaseDetail(item *unstructured.Unstructured) *HelmReleaseDetail {
	info := parseHelmReleaseInfo(item)

	chartName, _, _ := unstructured.NestedString(item.Object, "spec", "chart", "spec", "chart")
	chartVersion, _, _ := unstructured.NestedString(item.Object, "spec", "chart", "spec", "version")

	// 序列化 values
	valuesStr := ""
	if spec, ok := item.Object["spec"].(map[string]interface{}); ok {
		if values, ok := spec["values"]; ok {
			if valuesBytes, err := yaml.Marshal(values); err == nil {
				valuesStr = string(valuesBytes)
			}
		}
	}

	return &HelmReleaseDetail{
		HelmReleaseInfo: info,
		Values:          valuesStr,
		ChartName:       chartName,
		ChartVersion:    chartVersion,
		Notes:           "",
	}
}

func formatCRDStatus(status string) string {
	switch status {
	case "deployed":
		return "deployed"
	case "failed":
		return "failed"
	case "uninstalled":
		return "uninstalled"
	case "superseded":
		return "superseded"
	default:
		if status == "" {
			return "unknown"
		}
		return status
	}
}

func nestedInt(obj map[string]interface{}, keys ...string) int {
	current := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key]; ok {
				switch v := val.(type) {
				case float64:
					return int(v)
				case int:
					return v
				case json.Number:
					n, _ := v.Int64()
					return int(n)
				}
			}
			return 0
		}
		m, ok := current[key].(map[string]interface{})
		if !ok {
			return 0
		}
		current = m
	}
	return 0
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func fetchURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
