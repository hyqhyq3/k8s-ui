export interface HelmRelease {
  name: string
  namespace: string
  status: string
  chart: string
  appVersion: string
  revision: number
  updated: string
  description: string
}

export interface HelmReleaseDetail extends HelmRelease {
  values: string
  chartName: string
  chartVersion: string
  notes: string
}

export interface HelmReleaseHistory {
  revision: number
  chart: string
  appVersion: string
  status: string
  description: string
  updated: string
}

export interface HelmRepo {
  name: string
  url: string
}

export interface HelmChartSearchResult {
  name: string
  version: string
  appVersion: string
  description: string
  repo: string
}

export interface HelmChartVersion {
  version: string
  appVersion: string
  created: string
}

export interface HelmInstallRequest {
  name: string
  namespace: string
  chart: string
  repo?: string
  version?: string
  values?: string
  wait?: boolean
}

export interface HelmUpgradeRequest {
  chart?: string
  repo?: string
  version?: string
  values?: string
  wait?: boolean
  resetValues?: boolean
  reuseValues?: boolean
}

export interface HelmRollbackRequest {
  revision: number
}
