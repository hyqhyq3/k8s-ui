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
