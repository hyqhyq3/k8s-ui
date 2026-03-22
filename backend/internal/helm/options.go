package helm

import (
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"
)

// InstallOptions 安装选项
type InstallOptions struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Chart     string `json:"chart"`    // chart 名称或 URL
	Repo      string `json:"repo"`     // repo 名称（可选）
	Version   string `json:"version"`  // chart 版本（空=最新）
	Values    string `json:"values"`   // values YAML 字符串
	Wait      bool   `json:"wait"`
	Timeout   time.Duration
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
	Timeout     time.Duration
	ResetValues bool   `json:"resetValues"`
	ReuseValues bool   `json:"reuseValues"`
}

// RepoConfig Helm repo 配置文件路径
var settings = cli.New()

// ParseValues 解析 values YAML 字符串为 map
func (o *InstallOptions) ParseValues() (map[string]interface{}, error) {
	if o.Values == "" {
		return make(map[string]interface{}), nil
	}
	var values map[string]interface{}
	if err := yaml.Unmarshal([]byte(o.Values), &values); err != nil {
		return nil, fmt.Errorf("解析 values YAML 失败: %w", err)
	}
	return values, nil
}

// ParseValues 解析 values YAML 字符串为 map
func (o *UpgradeOptions) ParseValues() (map[string]interface{}, error) {
	if o.Values == "" {
		return make(map[string]interface{}), nil
	}
	var values map[string]interface{}
	if err := yaml.Unmarshal([]byte(o.Values), &values); err != nil {
		return nil, fmt.Errorf("解析 values YAML 失败: %w", err)
	}
	return values, nil
}

// ResolveChartPath 解析 chart 路径（支持 repo/chart 格式）
func (o *InstallOptions) ResolveChartPath(actionConfig *action.Configuration) (string, error) {
	if o.Repo != "" {
		chartRef := fmt.Sprintf("%s/%s", o.Repo, o.Chart)
		return resolveChartRef(chartRef, o.Version)
	}
	return resolveChartRef(o.Chart, o.Version)
}

// ResolveChartPath 解析 chart 路径
func (o *UpgradeOptions) ResolveChartPath(actionConfig *action.Configuration) (string, error) {
	if o.Repo != "" {
		chartRef := fmt.Sprintf("%s/%s", o.Repo, o.Chart)
		return resolveChartRef(chartRef, o.Version)
	}
	return resolveChartRef(o.Chart, o.Version)
}

// resolveChartRef 解析 chart 引用为本地路径
func resolveChartRef(chartRef, version string) (string, error) {
	install := action.NewInstall(&action.Configuration{})
	install.Version = version
	install.RepoURL = ""
	install.ChartPathOptions = action.ChartPathOptions{
		Version: version,
	}
	return install.LocateChart(chartRef, settings)
}

// LoadChart 加载本地 chart
func LoadChart(path string) (*loader.Chart, error) {
	return loader.Load(path)
}

// ParseValuesString 解析 --set 格式的字符串（如 key=value,key2=value2）
func ParseValuesString(str string) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	if err := strvals.ParseInto(str, values); err != nil {
		return nil, err
	}
	return values, nil
}

// GetRepoConfig 获取 Helm repo 配置文件
func GetRepoConfig() string {
	return settings.RepositoryConfig
}

// LoadRepoFile 加载 repo 配置文件
func LoadRepoFile() (*repo.File, error) {
	return repo.LoadFile(settings.RepositoryConfig)
}

// WriteRepoFile 写入 repo 配置文件
func WriteRepoFile(rf *repo.File) error {
	return rf.WriteFile(settings.RepositoryConfig, 0644)
}

// GetProviders 返回 Helm 使用的 getter providers
func GetProviders() getter.Providers {
	return getter.All(settings)
}
