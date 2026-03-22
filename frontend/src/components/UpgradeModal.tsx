import { useState, useEffect } from 'react'
import { Modal, Form, Input, Select, message } from 'antd'
import { upgradeRelease, fetchRepos, searchChart, fetchChartVersions } from '../api/helm'
import type { HelmRepo, HelmChartSearchResult, HelmChartVersion } from '../types/helm'

interface UpgradeModalProps {
  open: boolean
  namespace: string
  name: string
  currentChart: string
  currentVersion: string
  currentValues: string
  onClose: () => void
  onSuccess: () => void
}

export default function UpgradeModal({
  open, namespace, name, currentChart, currentVersion, currentValues, onClose, onSuccess,
}: UpgradeModalProps) {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [repos, setRepos] = useState<HelmRepo[]>([])
  const [selectedRepo, setSelectedRepo] = useState<string>('')
  const [chartOptions, setChartOptions] = useState<HelmChartSearchResult[]>([])
  const [versionOptions, setVersionOptions] = useState<HelmChartVersion[]>([])
  const [selectedChart, setSelectedChart] = useState<string>('')

  useEffect(() => {
    if (open) {
      fetchRepos().then(setRepos).catch(() => {})
      form.resetFields()
      form.setFieldValue('values', currentValues)
      setSelectedRepo('')
      setChartOptions([])
      setVersionOptions([])
      setSelectedChart('')
    }
  }, [open, currentValues])

  const handleSearchChart = async (keyword: string) => {
    if (!selectedRepo || !keyword) return
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

  const handleUpgrade = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      await upgradeRelease(namespace, name, {
        chart: values.chart || undefined,
        repo: values.repo || undefined,
        version: values.version || undefined,
        values: values.values || undefined,
      })

      message.success(`Release ${name} 升级成功`)
      onClose()
      onSuccess()
    } catch (err) {
      message.error(`升级失败: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title={`升级 ${name}`}
      open={open}
      onOk={handleUpgrade}
      onCancel={onClose}
      okText="升级"
      cancelText="取消"
      confirmLoading={loading}
      width={640}
    >
      <div style={{ marginBottom: 16, padding: '8px 12px', background: '#f6f8fa', borderRadius: 6, fontSize: 13 }}>
        当前: <strong>{currentChart}</strong> (chart v{currentVersion})
      </div>

      <Form form={form} layout="vertical">
        <Form.Item name="repo" label="Chart Repo（可选）">
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

        <Form.Item name="chart" label="Chart 名称（可选，留空保持当前）">
          <Select
            showSearch
            placeholder="搜索 chart..."
            disabled={!selectedRepo}
            filterOption={false}
            onSearch={handleSearchChart}
            onSelect={handleChartSelect}
            options={chartOptions.map((c) => ({
              label: `${c.name} (${c.version}) - ${c.description || 'No description'}`,
              value: c.name,
            }))}
          />
        </Form.Item>

        <Form.Item name="version" label="版本（留空使用最新）">
          <Select
            placeholder={selectedChart ? '选择版本' : '请先选择 chart'}
            allowClear
            disabled={!selectedChart}
            options={versionOptions.map((v) => ({
              label: `${v.version} (app: ${v.appVersion})`,
              value: v.version,
            }))}
          />
        </Form.Item>

        <Form.Item name="values" label="Values (YAML)">
          <Input.TextArea rows={10} placeholder="key: value" style={{ fontFamily: 'monospace' }} />
        </Form.Item>
      </Form>
    </Modal>
  )
}
