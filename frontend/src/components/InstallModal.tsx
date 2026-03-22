import { useState, useEffect } from 'react'
import { Modal, Form, Input, Select, message } from 'antd'
import { installRelease, fetchRepos, searchChart, fetchChartVersions } from '../api/helm'
import type { HelmRepo, HelmChartSearchResult, HelmChartVersion } from '../types/helm'

interface InstallModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export default function InstallModal({ open, onClose, onSuccess }: InstallModalProps) {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [repos, setRepos] = useState<HelmRepo[]>([])
  const [selectedRepo, setSelectedRepo] = useState<string>('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [chartOptions, setChartOptions] = useState<HelmChartSearchResult[]>([])
  const [versionOptions, setVersionOptions] = useState<HelmChartVersion[]>([])
  const [selectedChart, setSelectedChart] = useState<string>('')

  useEffect(() => {
    if (open) {
      fetchRepos().then(setRepos).catch(() => {})
      form.resetFields()
      setSelectedRepo('')
      setSearchKeyword('')
      setChartOptions([])
      setVersionOptions([])
      setSelectedChart('')
    }
  }, [open])

  const handleSearchChart = async (keyword: string) => {
    if (!selectedRepo || !keyword) return
    setSearchKeyword(keyword)
    try {
      const results = await searchChart(selectedRepo, keyword)
      setChartOptions(results)
    } catch {
      setChartOptions([])
    }
  }

  const handleChartSelect = async (chartName: string) => {
    setSelectedChart(chartName)
    form.setFieldValue('version', '')
    setVersionOptions([])
    if (!selectedRepo || !chartName) return
    try {
      const versions = await fetchChartVersions(selectedRepo, chartName)
      setVersionOptions(versions)
    } catch {
      // ignore
    }
  }

  const handleInstall = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      const repoName = values.repo
      const chartName = values.chart

      // 找到最新版本号
      let version = values.version
      if (!version) {
        const match = chartOptions.find((c) => c.name === chartName)
        if (match) version = match.version
      }

      await installRelease({
        name: values.name,
        namespace: values.namespace,
        chart: chartName,
        repo: repoName || undefined,
        version: version || undefined,
        values: values.values || undefined,
      })

      message.success(`Release ${values.name} 安装成功`)
      onClose()
      onSuccess()
    } catch (err) {
      message.error(`安装失败: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="安装 Helm Release"
      open={open}
      onOk={handleInstall}
      onCancel={onClose}
      okText="安装"
      cancelText="取消"
      confirmLoading={loading}
      width={640}
    >
      <Form form={form} layout="vertical" initialValues={{ wait: true }}>
        <Form.Item name="name" label="Release 名称" rules={[{ required: true }]}>
          <Input placeholder="例如: my-app" />
        </Form.Item>
        <Form.Item name="namespace" label="Namespace" rules={[{ required: true }]}>
          <Input placeholder="例如: default" />
        </Form.Item>

        <Form.Item name="repo" label="Chart Repo">
          <Select
            placeholder="选择 repo"
            allowClear
            onChange={(v) => {
              setSelectedRepo(v || '')
              setChartOptions([])
              setVersionOptions([])
              form.setFieldValue('chart', '')
            }}
            options={repos.map((r) => ({ label: `${r.name} (${r.url})`, value: r.name }))}
          />
        </Form.Item>

        <Form.Item name="chart" label="Chart 名称" rules={[{ required: true }]}>
          <Select
            showSearch
            placeholder={selectedRepo ? '搜索 chart...' : '请先选择 repo'}
            disabled={!selectedRepo}
            filterOption={false}
            onSearch={handleSearchChart}
            onSelect={handleChartSelect}
            options={chartOptions.map((c) => ({
              label: `${c.name} (${c.version}) - ${c.description || 'No description'}`,
              value: c.name,
            }))}
            notFoundContent={searchKeyword ? '未找到 chart' : '输入关键词搜索'}
          />
        </Form.Item>

        <Form.Item name="version" label="版本">
          <Select
            placeholder={selectedChart ? '选择版本（默认最新）' : '请先选择 chart'}
            allowClear
            disabled={!selectedChart}
            options={versionOptions.map((v) => ({
              label: `${v.version} (app: ${v.appVersion})`,
              value: v.version,
            }))}
          />
        </Form.Item>

        <Form.Item name="values" label="Values (YAML)">
          <Input.TextArea rows={6} placeholder="key: value" style={{ fontFamily: 'monospace' }} />
        </Form.Item>
      </Form>
    </Modal>
  )
}
