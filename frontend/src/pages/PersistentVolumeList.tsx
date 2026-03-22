import { Table, Tag, Space, Input, Button } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchPersistentVolumes } from '../api/k8s'
import type { PersistentVolume } from '../types/k8s'
import YAMLViewer from '../components/YAMLViewer'
import { useResourceList } from '../hooks/useResourceList'

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
  {
    title: '操作',
    key: 'actions',
    width: 60,
    render: (_: unknown, record: PersistentVolume) => (
      <YAMLViewer resourceType="pvs" name={record.name} />
    ),
  },
]

export default function PersistentVolumeList() {
  const { data, searchText, setSearchText, loading, refresh } =
    useResourceList<PersistentVolume>({ fetchFn: fetchPersistentVolumes })

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>PersistentVolumes</h2>
        <Space>
          <Input
            placeholder="搜索名称..."
            prefix={<SearchOutlined />}
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
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
