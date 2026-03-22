export interface Namespace {
  name: string
  status: string
  age: string
  labels: Record<string, string>
}

export interface Pod {
  name: string
  namespace: string
  status: string
  restarts: number
  node: string
  age: string
  ip: string
}

export interface Deployment {
  name: string
  namespace: string
  replicas: number
  ready: string
  age: string
  images: string[]
  labels: Record<string, string>
}

export interface StatefulSet {
  name: string
  namespace: string
  replicas: number
  ready: string
  age: string
  images: string[]
  labels: Record<string, string>
}

export interface DaemonSet {
  name: string
  namespace: string
  desired: number
  ready: number
  age: string
  images: string[]
  labels: Record<string, string>
}

export interface ConfigMap {
  name: string
  namespace: string
  keys: string[]
  age: string
  labels: Record<string, string>
}

export interface Secret {
  name: string
  namespace: string
  type: string
  age: string
  labels: Record<string, string>
}

export interface PersistentVolume {
  name: string
  capacity: string
  accessModes: string[]
  reclaimPolicy: string
  status: string
  storageClass: string
  claimRef: string
  age: string
}

export interface PersistentVolumeClaim {
  name: string
  namespace: string
  storageClass: string
  status: string
  volume: string
  accessModes: string[]
  capacity: string
  age: string
}

export interface StorageClass {
  name: string
  provisioner: string
  reclaimPolicy: string
  volumeBindingMode: string
  age: string
}

export interface ClusterStats {
  nodes: number
  namespaces: number
  pods: number
  deployments: number
  statefulSets: number
  daemonSets: number
  pvs: number
  pvcs: number
  version: string
  nodeStats: NodeStatInfo[]
}

export interface NodeStatInfo {
  name: string
  status: string
  pods: number
  podCapacity: number
  cpuAllocatable: string
  memoryAllocatable: string
}
