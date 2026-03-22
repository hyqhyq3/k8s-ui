import { useEffect, useState, useCallback } from 'react'
import { Table, Select, Space, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchConfigMaps, fetchNamespaces } from '../api/k8s'
import type { ConfigMap } from '../types/k8s'

const columns: ColumnsType<ConfigMap> = [
  {
    title: '名称',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
  },
  {
    title: 'Namespace',
    dataIndex: 'namespace',
    key: 'namespace',
  },
  {
    title: '数据键数',
    dataIndex: 'keys',
    key: 'keys',
    render: (keys: string[]) => keys.length,
    sorter: (a, b) => a.keys.length - b.keys.length,
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function ConfigMapList() {
  const [data, setData] = useState<ConfigMap[]>([])
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [selectedNs, setSelectedNs] = useState<string | undefined>(undefined)
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
      const res = await fetchConfigMaps(selectedNs)
      setData(res)
    } catch {
      message.error('获取 ConfigMap 列表失败')
    } finally {
      setLoading(false)
    }
  }, [selectedNs])

  useEffect(() => {
    loadNamespaces()
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>ConfigMaps</h2>
        <Select
          placeholder="按 Namespace 筛选"
          allowClear
          style={{ width: 240 }}
          value={selectedNs}
          onChange={setSelectedNs}
          options={namespaces.map((ns) => ({ label: ns, value: ns }))}
        />
      </div>
      <Table
        columns={columns}
        dataSource={data}
        rowKey={(row) => `${row.namespace}/${row.name}`}
        loading={loading}
        pagination={{ pageSize: 20 }}
      />
    </Space>
  )
}
