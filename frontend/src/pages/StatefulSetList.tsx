import { useEffect, useState, useCallback } from 'react'
import { Table, Select, Space, message, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchStatefulSets, fetchNamespaces } from '../api/k8s'
import type { StatefulSet } from '../types/k8s'

const columns: ColumnsType<StatefulSet> = [
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
    title: '就绪状态',
    dataIndex: 'ready',
    key: 'ready',
    render: (ready: string) => {
      const [readyCount, total] = ready.split('/').map(Number)
      return (
        <Tag color={readyCount === total ? 'green' : 'orange'}>
          {ready}
        </Tag>
      )
    },
  },
  {
    title: '副本数',
    dataIndex: 'replicas',
    key: 'replicas',
    sorter: (a, b) => a.replicas - b.replicas,
  },
  {
    title: '镜像',
    dataIndex: 'images',
    key: 'images',
    render: (images: string[]) => (
      <Space direction="vertical" size={2}>
        {images.map((img, i) => (
          <Tag key={i} style={{ fontSize: 12 }}>{img}</Tag>
        ))}
      </Space>
    ),
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function StatefulSetList() {
  const [data, setData] = useState<StatefulSet[]>([])
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
      const res = await fetchStatefulSets(selectedNs)
      setData(res)
    } catch {
      message.error('获取 StatefulSet 列表失败')
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
        <h2 style={{ margin: 0 }}>StatefulSets</h2>
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
