import { useState } from 'react'
import { Modal, Typography, Button, Spin, message, Tooltip } from 'antd'
import { CodeOutlined, CopyOutlined } from '@ant-design/icons'
import { fetchResourceYAML } from '../api/k8s'

const { Text } = Typography

interface YAMLViewerProps {
  resourceType: string
  name: string
  namespace?: string
}

export default function YAMLViewer({ resourceType, name, namespace }: YAMLViewerProps) {
  const [open, setOpen] = useState(false)
  const [yaml, setYaml] = useState('')
  const [loading, setLoading] = useState(false)

  const loadYAML = async () => {
    setLoading(true)
    try {
      const data = await fetchResourceYAML(resourceType, name, namespace)
      setYaml(data)
    } catch {
      message.error('获取 YAML 失败')
    } finally {
      setLoading(false)
    }
  }

  const handleOpen = () => {
    setOpen(true)
    loadYAML()
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(yaml)
      message.success('已复制到剪贴板')
    } catch {
      message.error('复制失败')
    }
  }

  return (
    <>
      <Tooltip title="查看 YAML">
        <Button type="link" size="small" icon={<CodeOutlined />} onClick={handleOpen} />
      </Tooltip>
      <Modal
        title={`YAML - ${name}`}
        open={open}
        onCancel={() => setOpen(false)}
        width={800}
        footer={[
          <Button key="copy" icon={<CopyOutlined />} onClick={handleCopy}>
            复制
          </Button>,
          <Button key="close" onClick={() => setOpen(false)}>
            关闭
          </Button>,
        ]}
      >
        {loading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
          </div>
        ) : (
          <pre
            style={{
              background: '#f5f5f5',
              padding: 16,
              borderRadius: 8,
              maxHeight: 500,
              overflow: 'auto',
              fontSize: 13,
              lineHeight: 1.6,
              margin: 0,
            }}
          >
            <Text code style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
              {yaml}
            </Text>
          </pre>
        )}
      </Modal>
    </>
  )
}
