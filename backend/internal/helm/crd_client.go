package helm

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured2 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var (
	HelmReleaseGVR = schema.GroupVersionResource{
		Group:    "helm.toolkit.fluxcd.io",
		Version:  "v2",
		Resource: "helmreleases",
	}
	HelmRepositoryGVR = schema.GroupVersionResource{
		Group:    "source.toolkit.fluxcd.io",
		Version:  "v1",
		Resource: "helmrepositories",
	}
	HelmReleaseGVK = schema.GroupVersionKind{
		Group:   "helm.toolkit.fluxcd.io",
		Version: "v2",
		Kind:    "HelmRelease",
	}
	HelmRepositoryGVK = schema.GroupVersionKind{
		Group:   "source.toolkit.fluxcd.io",
		Version: "v1",
		Kind:    "HelmRepository",
	}
)

// CRDClient wraps dynamic client for HelmRelease and HelmRepository CRDs
type CRDClient struct {
	dynamic dynamic.Interface
}

func NewCRDClient(config *rest.Config) (*CRDClient, error) {
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 dynamic client 失败: %w", err)
	}
	return &CRDClient{dynamic: dyn}, nil
}

// --- HelmRelease operations ---

// ListReleases lists HelmRelease CRDs across all namespaces or a specific one
func (c *CRDClient) ListReleases(ctx context.Context, namespace string) (*unstructured2.UnstructuredList, error) {
	if namespace == "" {
		// cluster-wide or all namespaces listing
		return c.dynamic.Resource(HelmReleaseGVR).List(ctx, metav1.ListOptions{})
	}
	return c.dynamic.Resource(HelmReleaseGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
}

// GetRelease gets a single HelmRelease CRD
func (c *CRDClient) GetRelease(ctx context.Context, namespace, name string) (*unstructured2.Unstructured, error) {
	return c.dynamic.Resource(HelmReleaseGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreateRelease creates a HelmRelease CRD
func (c *CRDClient) CreateRelease(ctx context.Context, namespace string, obj *unstructured2.Unstructured) (*unstructured2.Unstructured, error) {
	return c.dynamic.Resource(HelmReleaseGVR).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
}

// UpdateRelease updates a HelmRelease CRD
func (c *CRDClient) UpdateRelease(ctx context.Context, namespace string, obj *unstructured2.Unstructured) (*unstructured2.Unstructured, error) {
	return c.dynamic.Resource(HelmReleaseGVR).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
}

// DeleteRelease deletes a HelmRelease CRD (triggers helm uninstall by controller)
func (c *CRDClient) DeleteRelease(ctx context.Context, namespace, name string) error {
	return c.dynamic.Resource(HelmReleaseGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// RollbackRelease triggers rollback by patching the HelmRelease's chart version to a previous version
func (c *CRDClient) RollbackRelease(ctx context.Context, namespace, name string, targetRevision int) error {
	// Read current status.history to find the version at targetRevision
	hr, err := c.GetRelease(ctx, namespace, name)
	if err != nil {
		return err
	}
	history, found, _ := unstructured2.NestedSlice(hr.Object, "status", "history")
	_ = found
	if targetRevision < 1 || targetRevision > len(history) {
		return fmt.Errorf("revision %d 不在有效范围内 (1-%d)", targetRevision, len(history))
	}
	targetEntry, ok := history[targetRevision-1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无法读取 revision %d 的历史记录", targetRevision)
	}
	chartVersion, _, _ := unstructured2.NestedString(targetEntry, "chartVersion")
	if chartVersion == "" {
		return fmt.Errorf("revision %d 的 chartVersion 为空", targetRevision)
	}
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"chart": map[string]interface{}{
				"spec": map[string]interface{}{
					"version": chartVersion,
				},
			},
		},
	}
	patchBytes, _ := json.Marshal(patch)
	_, err = c.dynamic.Resource(HelmReleaseGVR).Namespace(namespace).Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	return err
}

// --- HelmRepository operations ---

// ListRepos lists HelmRepository CRDs in the flux-system namespace
func (c *CRDClient) ListRepos(ctx context.Context) (*unstructured2.UnstructuredList, error) {
	return c.dynamic.Resource(HelmRepositoryGVR).Namespace("flux-system").List(ctx, metav1.ListOptions{})
}

// CreateRepo creates a HelmRepository CRD
func (c *CRDClient) CreateRepo(ctx context.Context, obj *unstructured2.Unstructured) (*unstructured2.Unstructured, error) {
	return c.dynamic.Resource(HelmRepositoryGVR).Namespace("flux-system").Create(ctx, obj, metav1.CreateOptions{})
}

// DeleteRepo deletes a HelmRepository CRD
func (c *CRDClient) DeleteRepo(ctx context.Context, name string) error {
	return c.dynamic.Resource(HelmRepositoryGVR).Namespace("flux-system").Delete(ctx, name, metav1.DeleteOptions{})
}

// GetRepo gets a single HelmRepository CRD
func (c *CRDClient) GetRepo(ctx context.Context, name string) (*unstructured2.Unstructured, error) {
	return c.dynamic.Resource(HelmRepositoryGVR).Namespace("flux-system").Get(ctx, name, metav1.GetOptions{})
}
