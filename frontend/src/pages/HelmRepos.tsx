import { useEffect, useState, useCallback } from 'react'
import { Table, Button, Space, message, Modal, Form, Input, Popconfirm } from 'antd'
import { PlusOutlined, DeleteOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchRepos, addRepo, removeRepo, searchChart } from '../api/helm'
import type { HelmRepo, HelmChartSearchResult } from '../types/helm'

export default function HelmRepos() {
  const [repos, setRepos] = useState<HelmRepo[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()

  // 搜索
  const [searchRepo, setSearchRepo] = useState<string>('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchResults, setSearchResults] = useState<HelmChartSearchResult[]>([])
  const [searching, setSearching] = useState(false)

  const handleRemoveRepo = async (repoName: string) => {
    try {
      await removeRepo(repoName)
      message.success(`已删除 repo ${repoName}`)
      if (searchRepo === repoName) setSearchRepo('')
      loadData()
    } catch (err) {
      message.error(`删除 repo 失败: ${err}`)
    }
  }

  const handleAddRepo = async () => {
    try {
      const values = await form.validateFields()
      await addRepo(values.name, values.url)
      message.success(`已添加 repo ${values.name}`)
      setModalOpen(false)
      form.resetFields()
      loadData()
    } catch (err) {
      message.error(`添加 repo 失败: ${err}`)
    }
  }

  const handleSearch = async () => {
    if (!searchRepo || !searchKeyword) return
    setSearching(true)
    try {
      const results = await searchChart(searchRepo, searchKeyword)
      setSearchResults(results)
    } catch (err) {
      message.error(`搜索失败: ${err}`)
    } finally {
      setSearching(false)
    }
  }

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchRepos()
      setRepos(data)
      if (data.length > 0 && !searchRepo) {
        setSearchRepo(data[0].name)
      }
    } catch (err) {
      message.error(`加载 repo 列表失败: ${err}`)
    } finally {
      setLoading(false)
    }
  }, [searchRepo])

  useEffect(() => {
    loadData()
  }, [loadData])

  const repoColumns: ColumnsType<HelmRepo> = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Popconfirm
          title="确认删除"
          description={`删除 repo ${record.name}？`}
          onConfirm={() => handleRemoveRepo(record.name)}
        >
          <Button type="link" danger size="small" icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ]

  const searchColumns: ColumnsType<HelmChartSearchResult> = [
    { title: 'Chart', dataIndex: 'name', key: 'name' },
    { title: '版本', dataIndex: 'version', key: 'version', width: 100 },
    { title: 'AppVersion', dataIndex: 'appVersion', key: 'appVersion', width: 100 },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          添加 Repo
        </Button>
        <Button icon={<ReloadOutlined />} onClick={loadData} loading={loading}>
          刷新
        </Button>
      </Space>

      <Table
        rowKey="name"
        columns={repoColumns}
        dataSource={repos}
        loading={loading}
        pagination={false}
        size="small"
        style={{ marginBottom: 32 }}
      />

      {repos.length > 0 && (
        <>
          <h3>搜索 Chart</h3>
          <Space style={{ marginBottom: 16 }}>
            <select
              value={searchRepo}
              onChange={(e) => setSearchRepo(e.target.value)}
              style={{ width: 180, padding: '4px 8px', border: '1px solid #d9d9d9', borderRadius: 6 }}
            >
              {repos.map((r) => (
                <option key={r.name} value={r.name}>{r.name}</option>
              ))}
            </select>
            <Input
              placeholder="搜索关键词"
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              onPressEnter={handleSearch}
              style={{ width: 260 }}
              allowClear
            />
            <Button icon={<SearchOutlined />} onClick={handleSearch} loading={searching}>
              搜索
            </Button>
          </Space>

          <Table
            rowKey={(r) => `${r.repo}/${r.name}`}
            columns={searchColumns}
            dataSource={searchResults}
            loading={searching}
            pagination={{ pageSize: 20, showTotal: (t) => `共 ${t} 条` }}
            size="small"
          />
        </>
      )}

      <Modal
        title="添加 Chart Repo"
        open={modalOpen}
        onOk={handleAddRepo}
        onCancel={() => { setModalOpen(false); form.resetFields() }}
        okText="添加"
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入 repo 名称' }]}>
            <Input placeholder="例如: bitnami" />
          </Form.Item>
          <Form.Item name="url" label="URL" rules={[{ required: true, message: '请输入 repo URL' }]}>
            <Input placeholder="例如: https://charts.bitnami.com/bitnami" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
