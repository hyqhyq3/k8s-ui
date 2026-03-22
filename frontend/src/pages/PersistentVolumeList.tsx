import { useEffect, useState } from 'react'
import { Table, Space, message, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchPersistentVolumes } from '../api/k8s'
import type { PersistentVolume } from '../types/k8s'

const pvStatusColor: Record<string, string> = {
  Available: 'green',
  Bound: 'blue',
  Released: 'orange',
  Failed: 'red',
}

const columns: ColumnsType<PersistentVolume> = [
  {
    title: '名称',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
  },
  {
    title: '容量',
    dataIndex: 'capacity',
    key: 'capacity',
  },
  {
    title: '访问模式',
    dataIndex: 'accessModes',
    key: 'accessModes',
    render: (modes: string[]) => modes.join(', '),
  },
  {
    title: '回收策略',
    dataIndex: 'reclaimPolicy',
    key: 'reclaimPolicy',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) => (
      <Tag color={pvStatusColor[status] || 'default'}>{status}</Tag>
    ),
  },
  {
    title: 'StorageClass',
    dataIndex: 'storageClass',
    key: 'storageClass',
    render: (sc: string) => sc || '-',
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function PersistentVolumeList() {
  const [data, setData] = useState<PersistentVolume[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      try {
        const res = await fetchPersistentVolumes()
        setData(res)
      } catch {
        message.error('获取 PersistentVolume 列表失败')
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <h2 style={{ margin: 0 }}>PersistentVolumes</h2>
      <Table
        columns={columns}
        dataSource={data}
        rowKey="name"
        loading={loading}
        pagination={{ pageSize: 20 }}
      />
    </Space>
  )
}
