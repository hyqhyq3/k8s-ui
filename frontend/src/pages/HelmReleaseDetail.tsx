import { useEffect, useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Tabs, Table, Descriptions, Tag, Button, Space, Popconfirm, message, Spin } from 'antd'
import { ArrowLeftOutlined, RollbackOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import UpgradeModal from '../components/UpgradeModal'
import ValuesViewer from '../components/ValuesViewer'
import {
  fetchRelease,
  fetchReleaseHistory,
  fetchReleaseResources,
  uninstallRelease,
  rollbackRelease,
} from '../api/helm'
import type { HelmReleaseDetail, HelmReleaseHistory, HelmResource } from '../types/helm'

const statusColorMap: Record<string, string> = {
  deployed: 'green',
  failed: 'red',
  pending: 'orange',
  uninstalled: 'default',
  uninstalling: 'default',
  superseded: 'default',
}

const resourceColumns: ColumnsType<HelmResource> = [
  { title: 'Kind', dataIndex: 'kind', key: 'kind' },
  { title: 'Name', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: 'Namespace', dataIndex: 'namespace', key: 'namespace' },
  { title: 'API Version', dataIndex: 'apiVersion', key: 'apiVersion', ellipsis: true },
]

export default function HelmReleaseDetail() {
  const { namespace, name } = useParams<{ namespace: string; name: string }>()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<HelmReleaseDetail | null>(null)
  const [history, setHistory] = useState<HelmReleaseHistory[]>([])
  const [resources, setResources] = useState<HelmResource[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('info')
  const [upgradeOpen, setUpgradeOpen] = useState(false)

  const handleRollback = async (revision: number) => {
    if (!namespace || !name) return
    try {
      await rollbackRelease(namespace, name, { revision })
      message.success(`已回滚到 revision ${revision}`)
      loadData()
    } catch (err) {
      message.error(`回滚失败: ${err}`)
    }
  }

  const historyColumns: ColumnsType<HelmReleaseHistory> = useMemo(() => [
    { title: 'Revision', dataIndex: 'revision', key: 'revision', width: 90 },
    { title: 'Chart', dataIndex: 'chart', key: 'chart' },
    { title: 'AppVersion', dataIndex: 'appVersion', key: 'appVersion' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => <Tag color={statusColorMap[s] || 'default'}>{s}</Tag>,
    },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    { title: '更新时间', dataIndex: 'updated', key: 'updated', width: 180 },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: unknown, record: HelmReleaseHistory) => (
        <Popconfirm
          title="确认回滚"
          description={`回滚到 revision ${record.revision}？`}
          onConfirm={() => handleRollback(record.revision)}
        >
          <Button type="link" size="small" icon={<RollbackOutlined />}>
            回滚
          </Button>
        </Popconfirm>
      ),
    },
  ], [namespace, name])

  useEffect(() => {
    if (!namespace || !name) return
    loadData()
  }, [namespace, name])

  const loadData = async () => {
    if (!namespace || !name) return
    setLoading(true)
    try {
      const [detailData, historyData, resourceData] = await Promise.all([
        fetchRelease(namespace, name),
        fetchReleaseHistory(namespace, name).catch(() => []),
        fetchReleaseResources(namespace, name).catch(() => []),
      ])
      setDetail(detailData)
      setHistory(historyData)
      setResources(resourceData)
    } catch (err) {
      message.error(`加载 release 详情失败: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  const handleUninstall = async () => {
    if (!namespace || !name) return
    try {
      await uninstallRelease(namespace, name)
      message.success(`${name} 已卸载`)
      navigate('/helm/releases')
    } catch (err) {
      message.error(`卸载失败: ${err}`)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!detail) return <div>Release 不存在</div>

  const tabItems = [
    {
      key: 'info',
      label: '信息',
      children: (
        <Descriptions bordered column={2} size="small">
          <Descriptions.Item label="名称">{detail.name}</Descriptions.Item>
          <Descriptions.Item label="Namespace">{detail.namespace}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={statusColorMap[detail.status] || 'default'}>{detail.status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Revision">{detail.revision}</Descriptions.Item>
          <Descriptions.Item label="Chart">{detail.chartName}</Descriptions.Item>
          <Descriptions.Item label="Chart Version">{detail.chartVersion}</Descriptions.Item>
          <Descriptions.Item label="App Version">{detail.appVersion}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{detail.updated}</Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>{detail.description}</Descriptions.Item>
          {detail.notes && <Descriptions.Item label="Notes" span={2}><pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{detail.notes}</pre></Descriptions.Item>}
        </Descriptions>
      ),
    },
    {
      key: 'values',
      label: 'Values',
      children: <ValuesViewer content={detail.values || '{}'} />,
    },
    {
      key: 'resources',
      label: '资源',
      children: (
        <Table
          rowKey={(r) => `${r.kind}/${r.name}`}
          columns={resourceColumns}
          dataSource={resources}
          pagination={false}
          size="small"
        />
      ),
    },
    {
      key: 'history',
      label: `历史 (${history.length})`,
      children: (
        <Table
          rowKey="revision"
          columns={historyColumns}
          dataSource={history}
          pagination={false}
          size="small"
          scroll={{ x: 700 }}
        />
      ),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/helm/releases')}>
          返回
        </Button>
        <Popconfirm
          title="确认卸载"
          description={`确定要卸载 ${name} 吗？此操作不可逆。`}
          onConfirm={handleUninstall}
        >
          <Button danger icon={<DeleteOutlined />}>卸载</Button>
        </Popconfirm>
        <Button icon={<EditOutlined />} onClick={() => setUpgradeOpen(true)}>升级</Button>
      </Space>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        size="small"
      />

      <UpgradeModal
        open={upgradeOpen}
        namespace={detail.namespace}
        name={detail.name}
        currentChart={detail.chartName}
        currentVersion={detail.chartVersion}
        currentValues={detail.values}
        onClose={() => setUpgradeOpen(false)}
        onSuccess={loadData}
      />
    </div>
  )
}
