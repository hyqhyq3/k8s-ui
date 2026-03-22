import { get } from './client'
import type { Namespace, Pod, Deployment, StatefulSet, DaemonSet, ConfigMap, Secret, PersistentVolume, PersistentVolumeClaim, StorageClass } from '../types/k8s'

export function fetchNamespaces(): Promise<Namespace[]> {
  return get<Namespace[]>('/api/v1/namespaces')
}

export function fetchPods(namespace?: string): Promise<Pod[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<Pod[]>(`/api/v1/pods${params}`)
}

export function fetchDeployments(namespace?: string): Promise<Deployment[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<Deployment[]>(`/api/v1/deployments${params}`)
}

export function fetchStatefulSets(namespace?: string): Promise<StatefulSet[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<StatefulSet[]>(`/api/v1/statefulsets${params}`)
}

export function fetchDaemonSets(namespace?: string): Promise<DaemonSet[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<DaemonSet[]>(`/api/v1/daemonsets${params}`)
}

export function fetchConfigMaps(namespace?: string): Promise<ConfigMap[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<ConfigMap[]>(`/api/v1/configmaps${params}`)
}

export function fetchSecrets(namespace?: string): Promise<Secret[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<Secret[]>(`/api/v1/secrets${params}`)
}

export function fetchPersistentVolumes(): Promise<PersistentVolume[]> {
  return get<PersistentVolume[]>('/api/v1/pvs')
}

export function fetchPersistentVolumeClaims(namespace?: string): Promise<PersistentVolumeClaim[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<PersistentVolumeClaim[]>(`/api/v1/pvcs${params}`)
}

export function fetchStorageClasses(): Promise<StorageClass[]> {
  return get<StorageClass[]>('/api/v1/storageclasses')
}
