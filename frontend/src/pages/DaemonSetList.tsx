import { Table, Tag, Space, Input, Select, Button } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchDaemonSets } from '../api/k8s'
import type { DaemonSet } from '../types/k8s'
import YAMLViewer from '../components/YAMLViewer'
import { useResourceList } from '../hooks/useResourceList'

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
  {
    title: '操作',
    key: 'actions',
    width: 60,
    render: (_: unknown, record: DaemonSet) => (
      <YAMLViewer resourceType="daemonsets" name={record.name} namespace={record.namespace} />
    ),
  },
]

export default function DaemonSetList() {
  const { data, namespaces, selectedNs, setSelectedNs, searchText, setSearchText, loading, refresh } =
    useResourceList<DaemonSet>({ fetchFn: fetchDaemonSets })

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>DaemonSets</h2>
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
