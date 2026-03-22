import { Table, Tag, Space, Input, Select, Button } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchPods } from '../api/k8s'
import type { Pod } from '../types/k8s'
import YAMLViewer from '../components/YAMLViewer'
import { useResourceList } from '../hooks/useResourceList'

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
  {
    title: '操作',
    key: 'actions',
    width: 60,
    render: (_: unknown, record: Pod) => (
      <YAMLViewer resourceType="pods" name={record.name} namespace={record.namespace} />
    ),
  },
]

export default function PodList() {
  const { data, namespaces, selectedNs, setSelectedNs, searchText, setSearchText, loading, refresh } =
    useResourceList<Pod>({ fetchFn: fetchPods })

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>Pods</h2>
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
            placeholder="按 Namespace 筛选"
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
        rowKey={(row) => `${row.namespace}/${row.name}`}
        loading={loading}
        pagination={{ pageSize: 20, showTotal: (total) => `共 ${total} 条` }}
      />
    </Space>
  )
}
