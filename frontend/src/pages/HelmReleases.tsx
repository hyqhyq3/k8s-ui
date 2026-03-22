import { useEffect, useState, useCallback } from 'react'
import { Table, Tag, Select, Space, Button, Input, Popconfirm, message } from 'antd'
import { ReloadOutlined, DeleteOutlined, SearchOutlined, PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import type { ColumnsType } from 'antd/es/table'
import { fetchReleases, uninstallRelease } from '../api/helm'
import type { HelmRelease } from '../types/helm'
import { fetchNamespaces } from '../api/k8s'
import InstallModal from '../components/InstallModal'

const statusColorMap: Record<string, string> = {
  deployed: 'green',
  failed: 'red',
  pending: 'orange',
  uninstalled: 'default',
  uninstalling: 'default',
  superseded: 'default',
}

export default function HelmReleases() {
  const [releases, setReleases] = useState<HelmRelease[]>([])
  const [loading, setLoading] = useState(false)
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [selectedNs, setSelectedNs] = useState<string>('')
  const [searchText, setSearchText] = useState('')
  const [installOpen, setInstallOpen] = useState(false)
  const navigate = useNavigate()

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [data, ns] = await Promise.all([
        fetchReleases(selectedNs || undefined),
        fetchNamespaces(),
      ])
      setReleases(data)
      setNamespaces(ns.map((n: { name: string }) => n.name))
    } catch (err) {
      message.error(`加载 release 列表失败: ${err}`)
    } finally {
      setLoading(false)
    }
  }, [selectedNs])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handleUninstall = async (record: HelmRelease) => {
    try {
      await uninstallRelease(record.namespace, record.name)
      message.success(`${record.name} 已卸载`)
      loadData()
    } catch (err) {
      message.error(`卸载失败: ${err}`)
    }
  }

  const filteredReleases = releases.filter((r) => {
    if (!searchText) return true
    const s = searchText.toLowerCase()
    return r.name.toLowerCase().includes(s) ||
      r.namespace.toLowerCase().includes(s) ||
      r.chart.toLowerCase().includes(s)
  })

  const columns: ColumnsType<HelmRelease> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      sorter: (a, b) => a.name.localeCompare(b.name),
    },
    {
      title: 'Namespace',
      dataIndex: 'namespace',
      key: 'namespace',
      sorter: (a, b) => a.namespace.localeCompare(b.namespace),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 140,
      render: (status: string) => (
        <Tag color={statusColorMap[status] || 'default'}>{status}</Tag>
      ),
    },
    {
      title: 'Chart',
      key: 'chart',
      render: (_, r) => `${r.chart}-${r.appVersion}`,
    },
    {
      title: 'Revision',
      dataIndex: 'revision',
      key: 'revision',
      width: 90,
      sorter: (a, b) => a.revision - b.revision,
    },
    {
      title: '更新时间',
      dataIndex: 'updated',
      key: 'updated',
      width: 180,
      sorter: (a, b) => new Date(a.updated).getTime() - new Date(b.updated).getTime(),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record) => (
        <Popconfirm
          title="确认卸载"
          description={`确定要卸载 ${record.name} 吗？`}
          onConfirm={(e) => {
            e?.stopPropagation()
            handleUninstall(record)
          }}
          onCancel={(e) => e?.stopPropagation()}
        >
          <Button type="link" danger size="small" icon={<DeleteOutlined />}>
            卸载
          </Button>
        </Popconfirm>
      ),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setInstallOpen(true)}>
          安装
        </Button>
        <Input
          placeholder="搜索名称、namespace、chart"
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          style={{ width: 260 }}
          allowClear
        />
        <Select
          placeholder="Namespace"
          value={selectedNs || undefined}
          onChange={(v) => setSelectedNs(v || '')}
          allowClear
          style={{ width: 180 }}
          options={namespaces.map((ns) => ({ label: ns, value: ns }))}
        />
        <Button icon={<ReloadOutlined />} onClick={loadData} loading={loading}>
          刷新
        </Button>
      </Space>

      <Table
        rowKey={(r) => `${r.namespace}/${r.name}`}
        columns={columns}
        dataSource={filteredReleases}
        loading={loading}
        pagination={{ pageSize: 20, showSizeChanger: true, showTotal: (t) => `共 ${t} 条` }}
        onRow={(record) => ({
          onClick: () => navigate(`/helm/releases/${record.namespace}/${record.name}`),
          style: { cursor: 'pointer' },
        })}
      />

      <InstallModal
        open={installOpen}
        onClose={() => setInstallOpen(false)}
        onSuccess={loadData}
      />
    </div>
  )
}
