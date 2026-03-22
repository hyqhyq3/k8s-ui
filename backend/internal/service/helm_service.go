package service

import (
	"context"
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/yangqihuang/k8s-ui/internal/helm"
)

// HelmService Helm 业务逻辑
type HelmService struct {
	driver   *helm.Driver
}

// NewHelmService 创建 Helm 服务
func NewHelmService(driver *helm.Driver) *HelmService {
	return &HelmService{driver: driver}
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

// HelmResourceInfo release 管理的资源
type HelmResourceInfo struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
}

// HelmRepoInfo chart repo 信息
type HelmRepoInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ListReleases 列出 Helm release
func (s *HelmService) ListReleases(ctx context.Context, namespace string) ([]HelmReleaseInfo, error) {
	releases, err := s.driver.ListReleases(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取 release 列表失败: %w", err)
	}

	result := make([]HelmReleaseInfo, 0, len(releases))
	for _, r := range releases {
		result = append(result, HelmReleaseInfo{
			Name:        r.Name,
			Namespace:   r.Namespace,
			Status:      formatReleaseStatus(r.Info.Status),
			Chart:       r.Chart.Metadata.Name,
			AppVersion:  r.Chart.Metadata.AppVersion,
			Revision:    r.Version,
			Updated:     r.Info.LastDeployed.Time,
			Description: r.Info.Description,
		})
	}
	return result, nil
}

// GetRelease 获取 release 详情
func (s *HelmService) GetRelease(namespace, name string) (*HelmReleaseDetail, error) {
	r, err := s.driver.GetRelease(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取 release 详情失败: %w", err)
	}

	return &HelmReleaseDetail{
		HelmReleaseInfo: HelmReleaseInfo{
			Name:        r.Name,
			Namespace:   r.Namespace,
			Status:      formatReleaseStatus(r.Info.Status),
			Chart:       r.Chart.Metadata.Name,
			AppVersion:  r.Chart.Metadata.AppVersion,
			Revision:    r.Version,
			Updated:     r.Info.LastDeployed.Time,
			Description: r.Info.Description,
		},
		Values:       r.Config,
		ChartName:    r.Chart.Metadata.Name,
		ChartVersion: r.Chart.Metadata.Version,
		Notes:        r.Info.Notes,
	}, nil
}

// GetReleaseHistory 获取 release revision 历史
func (s *HelmService) GetReleaseHistory(namespace, name string) ([]HelmReleaseHistory, error) {
	releases, err := s.driver.GetReleaseHistory(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取 release 历史失败: %w", err)
	}

	result := make([]HelmReleaseHistory, 0, len(releases))
	for _, r := range releases {
		result = append(result, HelmReleaseHistory{
			Revision:    r.Version,
			Chart:       fmt.Sprintf("%s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version),
			AppVersion:  r.Chart.Metadata.AppVersion,
			Status:      formatReleaseStatus(r.Info.Status),
			Description: r.Info.Description,
			Updated:     r.Info.LastDeployed.Time,
		})
	}
	return result, nil
}

// GetReleaseResources 获取 release 管理的资源
func (s *HelmService) GetReleaseResources(namespace, name string) ([]HelmResourceInfo, error) {
	resources, err := s.driver.GetReleaseResources(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取 release 资源失败: %w", err)
	}

	result := make([]HelmResourceInfo, 0, len(resources))
	for _, res := range resources {
		result = append(result, HelmResourceInfo{
			APIVersion: res.APIVersion,
			Kind:       res.Kind,
			Name:       res.Name,
			Namespace:  res.Namespace,
		})
	}
	return result, nil
}

// UninstallRelease 卸载 release
func (s *HelmService) UninstallRelease(namespace, name string, keepHistory bool) error {
	_, err := s.driver.UninstallRelease(namespace, name, keepHistory)
	if err != nil {
		return fmt.Errorf("卸载 release 失败: %w", err)
	}
	return nil
}

// RollbackRelease 回滚 release
func (s *HelmService) RollbackRelease(namespace, name string, revision int) error {
	err := s.driver.RollbackRelease(namespace, name, revision)
	if err != nil {
		return fmt.Errorf("回滚 release 失败: %w", err)
	}
	return nil
}

// InstallRelease 安装 release
func (s *HelmService) InstallRelease(ctx context.Context, opts *helm.InstallOptions) (*HelmReleaseDetail, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Minute
	}

	r, err := s.driver.InstallRelease(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("安装 release 失败: %w", err)
	}

	return releaseToDetail(r), nil
}

// UpgradeRelease 升级 release
func (s *HelmService) UpgradeRelease(ctx context.Context, opts *helm.UpgradeOptions) (*HelmReleaseDetail, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Minute
	}

	r, err := s.driver.UpgradeRelease(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("升级 release 失败: %w", err)
	}

	return releaseToDetail(r), nil
}

// ListRepos 列出 chart repo
func (s *HelmService) ListRepos() ([]HelmRepoInfo, error) {
	rf, err := helm.LoadRepoFile()
	if err != nil {
		return nil, fmt.Errorf("加载 repo 配置失败: %w", err)
	}

	result := make([]HelmRepoInfo, 0, len(rf.Repositories))
	for _, r := range rf.Repositories {
		result = append(result, HelmRepoInfo{
			Name: r.Name,
			URL:  r.URL,
		})
	}
	return result, nil
}

// AddRepo 添加 chart repo
func (s *HelmService) AddRepo(name, url string) error {
	rf, err := helm.LoadRepoFile()
	if err != nil {
		rf = repo.NewFile()
	}

	if rf.Has(name) {
		return fmt.Errorf("repo %s 已存在", name)
	}

	rf.Add(&repo.Entry{
		Name: name,
		URL:  url,
	})

	if err := rf.WriteFile(helm.GetRepoConfig(), 0644); err != nil {
		return fmt.Errorf("写入 repo 配置失败: %w", err)
	}

	return nil
}

// RemoveRepo 删除 chart repo
func (s *HelmService) RemoveRepo(name string) error {
	rf, err := helm.LoadRepoFile()
	if err != nil {
		return fmt.Errorf("加载 repo 配置失败: %w", err)
	}

	if !rf.Has(name) {
		return fmt.Errorf("repo %s 不存在", name)
	}

	rf.Remove(name)
	if err := rf.WriteFile(helm.GetRepoConfig(), 0644); err != nil {
		return fmt.Errorf("写入 repo 配置失败: %w", err)
	}

	return nil
}

// SearchChart 在 repo 中搜索 chart
func (s *HelmService) SearchChart(repoName, keyword string) ([]map[string]interface{}, error) {
	rf, err := helm.LoadRepoFile()
	if err != nil {
		return nil, fmt.Errorf("加载 repo 配置失败: %w", err)
	}

	var repoEntry *repo.Entry
	for _, r := range rf.Repositories {
		if r.Name == repoName {
			repoEntry = r
			break
		}
	}
	if repoEntry == nil {
		return nil, fmt.Errorf("repo %s 不存在", repoName)
	}

	// 更新 repo index
	chartRepo, err := repo.NewChartRepository(repoEntry, helm.GetProviders())
	if err != nil {
		return nil, fmt.Errorf("创建 repo 引用失败: %w", err)
	}

	indexPath, err := chartRepo.DownloadIndexFile()
	if err != nil {
		return nil, fmt.Errorf("下载 repo index 失败: %w", err)
	}

	// 解析 index 文件搜索
	indexFile, err := repo.LoadIndexFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("加载 repo index 失败: %w", err)
	}

	results := make([]map[string]interface{}, 0)
	chartVersions, err := indexFile.Search(keyword)
	if err != nil {
		return nil, fmt.Errorf("搜索 chart 失败: %w", err)
	}

	for name, versions := range chartVersions {
		if len(versions) == 0 {
			continue
		}
		latest := versions[0]
		results = append(results, map[string]interface{}{
			"name":        name,
			"version":     latest.Version,
			"appVersion":  latest.AppVersion,
			"description": latest.Description,
			"repo":        repoName,
		})
	}
	return results, nil
}

// GetChartVersions 获取 chart 的可用版本列表
func (s *HelmService) GetChartVersions(repoName, chartName string) ([]map[string]string, error) {
	rf, err := helm.LoadRepoFile()
	if err != nil {
		return nil, fmt.Errorf("加载 repo 配置失败: %w", err)
	}

	var repoEntry *repo.Entry
	for _, r := range rf.Repositories {
		if r.Name == repoName {
			repoEntry = r
			break
		}
	}
	if repoEntry == nil {
		return nil, fmt.Errorf("repo %s 不存在", repoName)
	}

	chartRepo, err := repo.NewChartRepository(repoEntry, helm.GetProviders())
	if err != nil {
		return nil, fmt.Errorf("创建 repo 引用失败: %w", err)
	}

	indexPath, err := chartRepo.DownloadIndexFile()
	if err != nil {
		return nil, fmt.Errorf("下载 repo index 失败: %w", err)
	}

	indexFile, err := repo.LoadIndexFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("加载 repo index 失败: %w", err)
	}

	chartVersions, exists := indexFile.Entries[chartName]
	if !exists {
		return nil, fmt.Errorf("chart %s 在 repo %s 中不存在", chartName, repoName)
	}

	results := make([]map[string]string, 0, len(chartVersions))
	for _, v := range chartVersions {
		results = append(results, map[string]string{
			"version":    v.Version,
			"appVersion": v.AppVersion,
			"created":    v.Created.String(),
		})
	}
	return results, nil
}

// --- helpers ---

func releaseToDetail(r *release.Release) *HelmReleaseDetail {
	return &HelmReleaseDetail{
		HelmReleaseInfo: HelmReleaseInfo{
			Name:        r.Name,
			Namespace:   r.Namespace,
			Status:      formatReleaseStatus(r.Info.Status),
			Chart:       r.Chart.Metadata.Name,
			AppVersion:  r.Chart.Metadata.AppVersion,
			Revision:    r.Version,
			Updated:     r.Info.LastDeployed.Time,
			Description: r.Info.Description,
		},
		Values:       r.Config,
		ChartName:    r.Chart.Metadata.Name,
		ChartVersion: r.Chart.Metadata.Version,
		Notes:        r.Info.Notes,
	}
}

func formatReleaseStatus(status release.Status) string {
	switch status {
	case release.StatusDeployed:
		return "deployed"
	case release.StatusUninstalled:
		return "uninstalled"
	case release.StatusFailed:
		return "failed"
	case release.StatusPendingInstall:
		return "pending-install"
	case release.StatusPendingUpgrade:
		return "pending-upgrade"
	case release.StatusPendingRollback:
		return "pending-rollback"
	case release.StatusUninstalling:
		return "uninstalling"
	case release.StatusSuperseded:
		return "superseded"
	default:
		return string(status)
	}
}
