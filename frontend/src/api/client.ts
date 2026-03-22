interface ApiResponse<T> {
  data: T
}

export async function get<T>(url: string): Promise<T> {
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error(`请求失败: ${response.status} ${response.statusText}`)
  }
  const json: ApiResponse<T> = await response.json()
  return json.data
}
