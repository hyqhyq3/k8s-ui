package helm

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Driver 封装 Helm SDK 操作
type Driver struct {
	restConfig *rest.Config
	k8sClient  *kubernetes.Clientset
}

// NewDriver 创建 Helm Driver
func NewDriver(restConfig *rest.Config, k8sClient *kubernetes.Clientset) *Driver {
	return &Driver{
		restConfig: restConfig,
		k8sClient:  k8sClient,
	}
}

// newActionConfig 初始化指定 namespace 的 Helm action 配置
func (d *Driver) newActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(d.restConfig, namespace, "secrets", func(format string, v ...interface{}) {
		fmt.Printf("[helm] "+format+"\n", v...)
	}); err != nil {
		return nil, fmt.Errorf("初始化 Helm action 配置失败 (ns=%s): %w", namespace, err)
	}
	return actionConfig, nil
}

// ListReleases 列出所有或指定 namespace 的 Helm Release
func (d *Driver) ListReleases(ctx context.Context, namespace string) ([]*release.Release, error) {
	if namespace == "" {
		// 全部 namespace：遍历所有 namespace 收集 release
		return d.listAllReleases(ctx)
	}

	actionConfig, err := d.newActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	list := action.NewList(actionConfig)
	list.All = true
	list.StateMask = action.ListDeployed | action.ListUninstalled | action.ListPendingInstall | action.ListPendingUpgrade | action.ListPendingRollback | action.ListFailed | action.ListUninstalling | action.ListSuperseded

	return list.Run()
}

// listAllReleases 遍历所有 namespace 列出 release
func (d *Driver) listAllReleases(ctx context.Context) ([]*release.Release, error) {
	namespaces, err := d.k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 namespace 列表失败: %w", err)
	}

	var allReleases []*release.Release
	for _, ns := range namespaces.Items {
		actionConfig, err := d.newActionConfig(ns.Name)
		if err != nil {
			// 跳过无权限的 namespace
			continue
		}

		list := action.NewList(actionConfig)
		list.All = true
		list.StateMask = action.ListDeployed | action.ListUninstalled | action.ListPendingInstall | action.ListPendingUpgrade | action.ListPendingRollback | action.ListFailed | action.ListUninstalling | action.ListSuperseded

		releases, err := list.Run()
		if err != nil {
			continue
		}
		allReleases = append(allReleases, releases...)
	}
	return allReleases, nil
}

// GetRelease 获取指定 release 的详情
func (d *Driver) GetRelease(namespace, name string) (*release.Release, error) {
	actionConfig, err := d.newActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	get := action.NewGet(actionConfig)
	return get.Run(name)
}

// GetReleaseHistory 获取 release 的 revision 历史
func (d *Driver) GetReleaseHistory(namespace, name string) ([]*release.Release, error) {
	actionConfig, err := d.newActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	history := action.NewHistory(actionConfig)
	history.Max = 20
	return history.Run(name)
}

// GetReleaseResources 获取 release 管理的资源列表
func (d *Driver) GetReleaseResources(namespace, name string) ([]release.Resource, error) {
	actionConfig, err := d.newActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	get := action.NewGet(actionConfig)
	rel, err := get.Run(name)
	if err != nil {
		return nil, err
	}

	return rel.Resources(), nil
}

// UninstallRelease 卸载 release
func (d *Driver) UninstallRelease(namespace, name string, keepHistory bool) (*release.UninstallReleaseResponse, error) {
	actionConfig, err := d.newActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	uninstall := action.NewUninstall(actionConfig)
	uninstall.KeepHistory = keepHistory
	return uninstall.Run(name)
}

// RollbackRelease 回滚 release 到指定 revision
func (d *Driver) RollbackRelease(namespace, name string, revision int) error {
	actionConfig, err := d.newActionConfig(namespace)
	if err != nil {
		return err
	}

	rollback := action.NewRollback(actionConfig)
	rollback.Version = revision
	rollback.Wait = true
	return rollback.Run(name)
}

// InstallRelease 安装新 release
func (d *Driver) InstallRelease(ctx context.Context, opts *InstallOptions) (*release.Release, error) {
	actionConfig, err := d.newActionConfig(opts.Namespace)
	if err != nil {
		return nil, err
	}

	install := action.NewInstall(actionConfig)
	install.ReleaseName = opts.Name
	install.Namespace = opts.Namespace
	install.Version = opts.Version
	install.Wait = opts.Wait
	install.Timeout = opts.Timeout
	install.SkipCRDs = false
	install.Atomic = false

	chartPath, err := opts.ResolveChartPath(actionConfig)
	if err != nil {
		return nil, err
	}

	values, err := opts.ParseValues()
	if err != nil {
		return nil, err
	}

	return install.Run(chartPath, values)
}

// UpgradeRelease 升级 release
func (d *Driver) UpgradeRelease(ctx context.Context, opts *UpgradeOptions) (*release.Release, error) {
	actionConfig, err := d.newActionConfig(opts.Namespace)
	if err != nil {
		return nil, err
	}

	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Namespace = opts.Namespace
	upgrade.Version = opts.Version
	upgrade.Wait = opts.Wait
	upgrade.Timeout = opts.Timeout
	upgrade.ResetValues = opts.ResetValues
	upgrade.ReuseValues = opts.ReuseValues

	chartPath, err := opts.ResolveChartPath(actionConfig)
	if err != nil {
		return nil, err
	}

	values, err := opts.ParseValues()
	if err != nil {
		return nil, err
	}

	return upgrade.Run(opts.Name, chartPath, values)
}

// ListChartRepos 列出已添加的 chart repo
func (d *Driver) ListChartRepos(actionConfig *action.Configuration) ([]string, error) {
	list := action.NewRepoList(actionConfig)
	return list.Run()
}

// IsNotFound 判断错误是否为 release 不存在
func IsNotFound(err error) bool {
	return err != nil && (err == driver.ErrReleaseNotFound || err.Error() == driver.ErrReleaseNotFound.Error())
}
