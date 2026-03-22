import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  CloudServerOutlined,
  AppstoreOutlined,
} from '@ant-design/icons'

const { Sider, Content } = Layout

const menuItems = [
  {
    key: '/namespaces',
    icon: <CloudServerOutlined />,
    label: 'Namespaces',
  },
  {
    key: '/pods',
    icon: <AppstoreOutlined />,
    label: 'Pods',
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
