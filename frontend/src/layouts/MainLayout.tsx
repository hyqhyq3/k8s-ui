import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  DashboardOutlined,
  CloudServerOutlined,
  AppstoreOutlined,
  RocketOutlined,
  DatabaseOutlined,
  CloudOutlined,
  FileTextOutlined,
  LockOutlined,
  HddOutlined,
  CloudSyncOutlined,
} from '@ant-design/icons'

const { Sider, Content } = Layout

const menuItems = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: '概览',
  },
  {
    key: '/namespaces',
    icon: <CloudServerOutlined />,
    label: 'Namespaces',
  },
  {
    key: '/workloads',
    icon: <AppstoreOutlined />,
    label: '工作负载',
    children: [
      {
        key: '/pods',
        icon: <AppstoreOutlined />,
        label: 'Pods',
      },
      {
        key: '/deployments',
        icon: <RocketOutlined />,
        label: 'Deployments',
      },
      {
        key: '/statefulsets',
        icon: <DatabaseOutlined />,
        label: 'StatefulSets',
      },
      {
        key: '/daemonsets',
        icon: <CloudOutlined />,
        label: 'DaemonSets',
      },
    ],
  },
  {
    key: '/config',
    icon: <FileTextOutlined />,
    label: '配置',
    children: [
      {
        key: '/configmaps',
        icon: <FileTextOutlined />,
        label: 'ConfigMaps',
      },
      {
        key: '/secrets',
        icon: <LockOutlined />,
        label: 'Secrets',
      },
    ],
  },
  {
    key: '/storage',
    icon: <HddOutlined />,
    label: '存储',
    children: [
      {
        key: '/pvs',
        icon: <HddOutlined />,
        label: 'PVs',
      },
      {
        key: '/pvcs',
        icon: <DatabaseOutlined />,
        label: 'PVCs',
      },
      {
        key: '/storageclasses',
        icon: <CloudServerOutlined />,
        label: 'StorageClasses',
      },
    ],
  },
  {
    key: '/helm',
    icon: <CloudSyncOutlined />,
    label: 'Helm',
    children: [
      {
        key: '/helm/releases',
        icon: <CloudSyncOutlined />,
        label: 'Releases',
      },
      {
        key: '/helm/repos',
        icon: <DatabaseOutlined />,
        label: 'Chart 仓库',
      },
    ],
  },
]

export default function MainLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ height: 32, margin: 16, color: '#fff', textAlign: 'center', fontSize: collapsed ? 14 : 18, fontWeight: 'bold' }}>
          {collapsed ? 'K8s' : 'K8s UI'}
        </div>
        <Menu
          theme="dark"
          selectedKeys={[location.pathname]}
          defaultOpenKeys={['/workloads', '/config', '/storage', '/helm']}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Content style={{ margin: 16 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
