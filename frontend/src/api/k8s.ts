import { get } from './client'
import type { Namespace, Pod } from '../types/k8s'

export function fetchNamespaces(): Promise<Namespace[]> {
  return get<Namespace[]>('/api/v1/namespaces')
}

export function fetchPods(namespace?: string): Promise<Pod[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<Pod[]>(`/api/v1/pods${params}`)
}
