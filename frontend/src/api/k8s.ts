import { get } from './client'
import type { Namespace, Pod, Deployment, StatefulSet, DaemonSet } from '../types/k8s'

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
