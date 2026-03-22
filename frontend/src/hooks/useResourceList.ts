import { useState, useEffect, useCallback } from 'react'
import { message } from 'antd'
import { fetchNamespaces } from '../api/k8s'

interface UseResourceListOptions<T> {
  fetchFn: (namespace?: string) => Promise<T[]>
  defaultNamespace?: string
}

export function useResourceList<T extends { name: string; namespace?: string }>(
  options: UseResourceListOptions<T>
) {
  const { fetchFn } = options
  const [data, setData] = useState<T[]>([])
  const [filteredData, setFilteredData] = useState<T[]>([])
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [selectedNs, setSelectedNs] = useState<string | undefined>(undefined)
  const [searchText, setSearchText] = useState('')
  const [loading, setLoading] = useState(false)

  const loadNamespaces = async () => {
    try {
      const res = await fetchNamespaces()
      setNamespaces(res.map((ns) => ns.name))
    } catch {
      message.error('获取 Namespace 列表失败')
    }
  }

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetchFn(selectedNs)
      setData(res)
    } catch {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }, [fetchFn, selectedNs])

  // 加载 namespace 列表
  useEffect(() => {
    loadNamespaces()
  }, [])

  // 加载数据
  useEffect(() => {
    loadData()
  }, [loadData])

  // 客户端搜索过滤
  useEffect(() => {
    if (!searchText) {
      setFilteredData(data)
      return
    }
    const lower = searchText.toLowerCase()
    setFilteredData(
      data.filter((item) => {
        return (
          item.name.toLowerCase().includes(lower) ||
          (item.namespace && item.namespace.toLowerCase().includes(lower))
        )
      })
    )
  }, [searchText, data])

  const refresh = () => {
    loadData()
  }

  return {
    data: filteredData,
    totalData: data,
    namespaces,
    selectedNs,
    setSelectedNs,
    searchText,
    setSearchText,
    loading,
    refresh,
  }
}
