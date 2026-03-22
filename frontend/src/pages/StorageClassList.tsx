import { Table, Tag, Space, Input, Button } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchStorageClasses } from '../api/k8s'
import type { StorageClass } from '../types/k8s'
import YAMLViewer from '../components/YAMLViewer'
import { useResourceList } from '../hooks/useResourceList'

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
  {
    title: '操作',
    key: 'actions',
    width: 60,
    render: (_: unknown, record: StorageClass) => (
      <YAMLViewer resourceType="storageclasses" name={record.name} />
    ),
  },
]

export default function StorageClassList() {
  const { data, searchText, setSearchText, loading, refresh } =
    useResourceList<StorageClass>({ fetchFn: fetchStorageClasses })

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>StorageClasses</h2>
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
