import { useEffect, useState, useCallback } from 'react'
import { Table, Select, Space, message, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchPersistentVolumeClaims, fetchNamespaces } from '../api/k8s'
import type { PersistentVolumeClaim } from '../types/k8s'

const pvcStatusColor: Record<string, string> = {
  Bound: 'green',
  Pending: 'orange',
  Lost: 'red',
}

const columns: ColumnsType<PersistentVolumeClaim> = [
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
    title: 'StorageClass',
    dataIndex: 'storageClass',
    key: 'storageClass',
    render: (sc: string) => sc || '-',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) => (
      <Tag color={pvcStatusColor[status] || 'default'}>{status}</Tag>
    ),
  },
  {
    title: 'Volume',
    dataIndex: 'volume',
    key: 'volume',
    render: (vol: string) => vol || '-',
  },
  {
    title: '容量',
    dataIndex: 'capacity',
    key: 'capacity',
    render: (cap: string) => cap || '-',
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function PersistentVolumeClaimList() {
  const [data, setData] = useState<PersistentVolumeClaim[]>([])
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
      const res = await fetchPersistentVolumeClaims(selectedNs)
      setData(res)
    } catch {
      message.error('获取 PersistentVolumeClaim 列表失败')
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
        <h2 style={{ margin: 0 }}>PersistentVolumeClaims</h2>
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
