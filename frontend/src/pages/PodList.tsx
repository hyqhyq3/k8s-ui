import { useEffect, useState, useCallback } from 'react'
import { Table, Tag, Select, Space, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchPods, fetchNamespaces } from '../api/k8s'
import type { Pod } from '../types/k8s'

const statusColorMap: Record<string, string> = {
  Running: 'green',
  Pending: 'orange',
  Succeeded: 'blue',
  Failed: 'red',
  Terminating: 'default',
  Unknown: 'default',
}

const columns: ColumnsType<Pod> = [
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
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) => (
      <Tag color={statusColorMap[status] || 'default'}>{status}</Tag>
    ),
  },
  {
    title: '重启次数',
    dataIndex: 'restarts',
    key: 'restarts',
    sorter: (a, b) => a.restarts - b.restarts,
  },
  {
    title: 'Node',
    dataIndex: 'node',
    key: 'node',
    ellipsis: true,
  },
  {
    title: 'Pod IP',
    dataIndex: 'ip',
    key: 'ip',
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function PodList() {
  const [pods, setPods] = useState<Pod[]>([])
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [selectedNs, setSelectedNs] = useState<string | undefined>(undefined)
  const [loading, setLoading] = useState(false)

  const loadNamespaces = async () => {
    try {
      const data = await fetchNamespaces()
      setNamespaces(data.map((ns) => ns.name))
    } catch {
      message.error('获取 Namespace 列表失败')
    }
  }

  const loadPods = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchPods(selectedNs)
      setPods(data)
    } catch {
      message.error('获取 Pod 列表失败')
    } finally {
      setLoading(false)
    }
  }, [selectedNs])

  useEffect(() => {
    loadNamespaces()
  }, [])

  useEffect(() => {
    loadPods()
  }, [loadPods])

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>Pods</h2>
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
        dataSource={pods}
        rowKey={(row) => `${row.namespace}/${row.name}`}
        loading={loading}
        pagination={{ pageSize: 20 }}
      />
    </Space>
  )
}
