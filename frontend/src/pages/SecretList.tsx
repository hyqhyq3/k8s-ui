import { useEffect, useState, useCallback } from 'react'
import { Table, Select, Space, message, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { fetchSecrets, fetchNamespaces } from '../api/k8s'
import type { Secret } from '../types/k8s'

const secretTypeColor: Record<string, string> = {
  'Opaque': 'default',
  'kubernetes.io/tls': 'blue',
  'kubernetes.io/dockerconfigjson': 'orange',
  'kubernetes.io/basic-auth': 'green',
  'kubernetes.io/ssh-auth': 'purple',
  'kubernetes.io/service-account-token': 'cyan',
}

const columns: ColumnsType<Secret> = [
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
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    render: (type: string) => (
      <Tag color={secretTypeColor[type] || 'default'}>{type}</Tag>
    ),
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
]

export default function SecretList() {
  const [data, setData] = useState<Secret[]>([])
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
      const res = await fetchSecrets(selectedNs)
      setData(res)
    } catch {
      message.error('获取 Secret 列表失败')
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
        <h2 style={{ margin: 0 }}>Secrets</h2>
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
