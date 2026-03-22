import { get, post, del } from './client'
import type {
  HelmRelease,
  HelmReleaseDetail,
  HelmReleaseHistory,
  HelmResource,
  HelmRepo,
  HelmChartSearchResult,
  HelmChartVersion,
  HelmInstallRequest,
  HelmUpgradeRequest,
  HelmRollbackRequest,
} from '../types/helm'

// --- Releases ---

export function fetchReleases(namespace?: string): Promise<HelmRelease[]> {
  const params = namespace ? `?namespace=${encodeURIComponent(namespace)}` : ''
  return get<HelmRelease[]>(`/api/v1/helm/releases${params}`)
}

export function fetchRelease(namespace: string, name: string): Promise<HelmReleaseDetail> {
  return get<HelmReleaseDetail>(`/api/v1/helm/releases/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`)
}

export function fetchReleaseHistory(namespace: string, name: string): Promise<HelmReleaseHistory[]> {
  return get<HelmReleaseHistory[]>(`/api/v1/helm/releases/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/history`)
}

export function fetchReleaseResources(namespace: string, name: string): Promise<HelmResource[]> {
  return get<HelmResource[]>(`/api/v1/helm/releases/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/resources`)
}

export function uninstallRelease(namespace: string, name: string, keepHistory = false): Promise<void> {
  const params = keepHistory ? '?keepHistory=true' : ''
  return del<void>(`/api/v1/helm/releases/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}${params}`)
}

export function rollbackRelease(namespace: string, name: string, data: HelmRollbackRequest): Promise<void> {
  return post<void>(`/api/v1/helm/releases/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/rollback`, data)
}

export function installRelease(data: HelmInstallRequest): Promise<HelmReleaseDetail> {
  return post<HelmReleaseDetail>('/api/v1/helm/install', data)
}

export function upgradeRelease(namespace: string, name: string, data: HelmUpgradeRequest): Promise<HelmReleaseDetail> {
  return post<HelmReleaseDetail>(`/api/v1/helm/releases/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/upgrade`, data)
}

// --- Repos ---

export function fetchRepos(): Promise<HelmRepo[]> {
  return get<HelmRepo[]>('/api/v1/helm/repos')
}

export function addRepo(name: string, url: string): Promise<void> {
  return post<void>('/api/v1/helm/repos', { name, url })
}

export function removeRepo(name: string): Promise<void> {
  return del<void>(`/api/v1/helm/repos/${encodeURIComponent(name)}`)
}

export function searchChart(repo: string, keyword: string): Promise<HelmChartSearchResult[]> {
  return get<HelmChartSearchResult[]>(`/api/v1/helm/repos/${encodeURIComponent(repo)}/search?q=${encodeURIComponent(keyword)}`)
}

export function fetchChartVersions(repo: string, chart: string): Promise<HelmChartVersion[]> {
  return get<HelmChartVersion[]>(`/api/v1/helm/repos/${encodeURIComponent(repo)}/charts/${encodeURIComponent(chart)}/versions`)
}
