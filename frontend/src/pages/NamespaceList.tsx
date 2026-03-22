import { Table, Space, Input, Select, Button } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchNamespaces } from '../api/k8s'
import type { Namespace } from '../types/k8s'
import YAMLViewer from '../components/YAMLViewer'
import { useResourceList } from '../hooks/useResourceList'

const columns: ColumnsType<Namespace> = [
  {
    title: '名称',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) => {
      const color = status === 'Active' ? 'green' : 'default'
      return <span style={{ color: color === 'green' ? '#52c41a' : undefined }}>{status}</span>
    },
  },
  {
    title: '创建时间',
    dataIndex: 'age',
    key: 'age',
  },
  {
    title: '操作',
    key: 'actions',
    width: 60,
    render: (_: unknown, record: Namespace) => (
      <YAMLViewer resourceType="namespaces" name={record.name} />
    ),
  },
]

export default function NamespaceList() {
  const { data, namespaces, selectedNs, setSelectedNs, searchText, setSearchText, loading, refresh } =
    useResourceList({ fetchFn: fetchNamespaces })

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>Namespaces</h2>
        <Space>
          <Input
            placeholder="搜索名称..."
            prefix={<SearchOutlined />}
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
          />
          <Select
            placeholder="筛选 Namespace"
            allowClear
            style={{ width: 200 }}
            value={selectedNs}
            onChange={setSelectedNs}
            options={namespaces.map((ns) => ({ label: ns, value: ns }))}
          />
          <Button icon={<ReloadOutlined />} onClick={refresh} loading={loading}>
            刷新
          </Button>
        </Space>
      </div>
      <Table
        columns={columns}
        dataSource={data}
        rowKey="name"
        loading={loading}
        pagination={{ pageSize: 20, showTotal: (total) => `共 ${total} 条` }}
      />
    </Space>
  )
}
