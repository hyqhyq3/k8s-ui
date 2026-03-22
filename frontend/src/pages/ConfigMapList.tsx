import { Table, Space, Input, Select, Button } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchConfigMaps } from '../api/k8s'
import type { ConfigMap } from '../types/k8s'
import YAMLViewer from '../components/YAMLViewer'
import { useResourceList } from '../hooks/useResourceList'

const columns: ColumnsType<ConfigMap> = [
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
    title: '数据键数',
    dataIndex: 'keys',
    key: 'keys',
    render: (keys: string[]) => keys.length,
    sorter: (a, b) => a.keys.length - b.keys.length,
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
    render: (_: unknown, record: ConfigMap) => (
      <YAMLViewer resourceType="configmaps" name={record.name} namespace={record.namespace} />
    ),
  },
]

export default function ConfigMapList() {
  const { data, namespaces, selectedNs, setSelectedNs, searchText, setSearchText, loading, refresh } =
    useResourceList<ConfigMap>({ fetchFn: fetchConfigMaps })

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>ConfigMaps</h2>
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
