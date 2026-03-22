import { useEffect, useState } from 'react'
import { Table, Space, message, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchStorageClasses } from '../api/k8s'
import type { StorageClass } from '../types/k8s'

const bindingModeColor: Record<string, string> = {
  Immediate: 'blue',
  WaitForFirstConsumer: 'orange',
}

const columns: ColumnsType<StorageClass> = [
  {
    title: '名称',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
  },
  {
    title: 'Provisioner',
    dataIndex: 'provisioner',
    key: 'provisioner',
    ellipsis: true,
  },
  {
    title: '回收策略',
    dataIndex: 'reclaimPolicy',
    key: 'reclaimPolicy',
    render: (policy: string) => policy || '-',
  },
  {
    title: '绑定模式',
    dataIndex: 'volumeBindingMode',
    key: 'volumeBindingMode',
    render: (mode: string) => (
      mode ? <Tag color={bindingModeColor[mode] || 'default'}>{mode}</Tag> : '-'
    ),
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function StorageClassList() {
  const [data, setData] = useState<StorageClass[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      try {
        const res = await fetchStorageClasses()
        setData(res)
      } catch {
        message.error('获取 StorageClass 列表失败')
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <h2 style={{ margin: 0 }}>StorageClasses</h2>
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
