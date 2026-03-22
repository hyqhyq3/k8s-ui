interface ApiResponse<T> {
  data: T
}

export async function get<T>(url: string): Promise<T> {
  const response = await fetch(url)
  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || `请求失败: ${response.status} ${response.statusText}`)
  }
  const json: ApiResponse<T> = await response.json()
  return json.data
}

export async function post<T>(url: string, body?: unknown): Promise<T> {
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || `请求失败: ${response.status} ${response.statusText}`)
  }
  const json: ApiResponse<T> = await response.json()
  return json.data
}

export async function del<T>(url: string): Promise<T> {
  const response = await fetch(url, { method: 'DELETE' })
  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || `请求失败: ${response.status} ${response.statusText}`)
  }
  const json: ApiResponse<T> = await response.json()
  return json.data
}
