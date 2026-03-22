import { useEffect, useState } from 'react'
import { Table, Tag, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchNamespaces } from '../api/k8s'
import type { Namespace } from '../types/k8s'

const columns: ColumnsType<Namespace> = [
  {
    title: '名称',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) => {
      const color = status === 'Active' ? 'green' : 'default'
      return <Tag color={color}>{status}</Tag>
    },
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function NamespaceList() {
  const [namespaces, setNamespaces] = useState<Namespace[]>([])
  const [loading, setLoading] = useState(false)

  const loadData = async () => {
    setLoading(true)
    try {
      const data = await fetchNamespaces()
      setNamespaces(data)
    } catch {
      message.error('获取 Namespace 列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  return (
    <Table
      title={() => <h2>Namespaces</h2>}
      columns={columns}
      dataSource={namespaces}
      rowKey="name"
      loading={loading}
      pagination={false}
    />
  )
}
