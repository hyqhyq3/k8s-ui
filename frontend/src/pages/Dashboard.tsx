import { useEffect, useState } from 'react'
import { Card, Col, Row, Statistic, Table, Tag, Spin, message } from 'antd'
import {
  CloudServerOutlined,
  AppstoreOutlined,
  RocketOutlined,
  DatabaseOutlined,
  CloudOutlined,
  HddOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fetchClusterStats } from '../api/k8s'
import type { ClusterStats, NodeStatInfo } from '../types/k8s'

const columns: ColumnsType<NodeStatInfo> = [
  {
    title: '节点名称',
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) => (
      <Tag color={status === 'Ready' ? 'green' : 'red'}>{status}</Tag>
    ),
  },
  {
    title: 'CPU',
    dataIndex: 'cpuAllocatable',
    key: 'cpuAllocatable',
  },
  {
    title: '内存',
    dataIndex: 'memoryAllocatable',
    key: 'memoryAllocatable',
  },
  {
    title: 'Pod 容量',
    dataIndex: 'podCapacity',
    key: 'podCapacity',
  },
]

export default function Dashboard() {
  const [stats, setStats] = useState<ClusterStats | null>(null)
  const [loading, setLoading] = useState(false)

  const loadStats = async () => {
    setLoading(true)
    try {
      const data = await fetchClusterStats()
      setStats(data)
    } catch {
      message.error('获取集群信息失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadStats()
  }, [])

  if (loading && !stats) {
    return (
      <div style={{ textAlign: 'center', padding: 80 }}>
        <Spin size="large" tip="加载集群信息..." />
      </div>
    )
  }

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>
          集群概览 {stats && <Tag color="blue">v{stats.version}</Tag>}
        </h2>
        <ReloadOutlined
          style={{ fontSize: 18, cursor: 'pointer', color: '#1677ff' }}
          onClick={loadStats}
          spin={loading}
        />
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="节点"
              value={stats?.nodes ?? 0}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="Namespaces"
              value={stats?.namespaces ?? 0}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="Pods"
              value={stats?.pods ?? 0}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="Deployments"
              value={stats?.deployments ?? 0}
              prefix={<RocketOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="StatefulSets"
              value={stats?.statefulSets ?? 0}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#13c2c2' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="DaemonSets"
              value={stats?.daemonSets ?? 0}
              prefix={<CloudOutlined />}
              valueStyle={{ color: '#eb2f96' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="PVs"
              value={stats?.pvs ?? 0}
              prefix={<HddOutlined />}
              valueStyle={{ color: '#2f54eb' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6}>
          <Card>
            <Statistic
              title="PVCs"
              value={stats?.pvcs ?? 0}
              prefix={<HddOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      {stats && stats.nodeStats && stats.nodeStats.length > 0 && (
        <Card title="节点状态" style={{ marginTop: 16 }}>
          <Table
            columns={columns}
            dataSource={stats.nodeStats}
            rowKey="name"
            pagination={false}
            size="middle"
          />
        </Card>
      )}
    </div>
  )
}
