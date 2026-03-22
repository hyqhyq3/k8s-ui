import { useEffect, useState, useCallback } from 'react'
import { Table, Select, Space, message, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchDaemonSets, fetchNamespaces } from '../api/k8s'
import type { DaemonSet } from '../types/k8s'

const columns: ColumnsType<DaemonSet> = [
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
    key: 'ready',
    render: (_: unknown, record: DaemonSet) => {
      const ready = record.desired === 0 || record.ready === record.desired
      return (
        <Tag color={ready ? 'green' : 'orange'}>
          {record.ready}/{record.desired}
        </Tag>
      )
    },
  },
  {
    title: '期望节点数',
    dataIndex: 'desired',
    key: 'desired',
    sorter: (a, b) => a.desired - b.desired,
  },
  {
    title: '就绪节点数',
    dataIndex: 'ready',
    key: 'readyNum',
    sorter: (a, b) => a.ready - b.ready,
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

export default function DaemonSetList() {
  const [data, setData] = useState<DaemonSet[]>([])
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
      const res = await fetchDaemonSets(selectedNs)
      setData(res)
    } catch {
      message.error('获取 DaemonSet 列表失败')
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
        <h2 style={{ margin: 0 }}>DaemonSets</h2>
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
