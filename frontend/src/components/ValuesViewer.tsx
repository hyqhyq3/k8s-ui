import { Button, message, Tooltip } from 'antd'
import { CopyOutlined } from '@ant-design/icons'

interface ValuesViewerProps {
  content: string
}

export default function ValuesViewer({ content }: ValuesViewerProps) {
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(content)
      message.success('已复制到剪贴板')
    } catch {
      message.error('复制失败')
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 8, textAlign: 'right' }}>
        <Tooltip title="复制 Values">
          <Button size="small" icon={<CopyOutlined />} onClick={handleCopy}>
            复制
          </Button>
        </Tooltip>
      </div>
      <pre
        style={{
          background: '#f5f5f5',
          padding: 16,
          borderRadius: 8,
          maxHeight: 600,
          overflow: 'auto',
          fontSize: 13,
          lineHeight: 1.6,
          margin: 0,
        }}
      >
        {content}
      </pre>
    </div>
  )
}
