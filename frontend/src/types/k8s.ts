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
