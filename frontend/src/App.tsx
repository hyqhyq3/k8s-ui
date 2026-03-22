import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import MainLayout from './layouts/MainLayout'
import Dashboard from './pages/Dashboard'
import NamespaceList from './pages/NamespaceList'
import PodList from './pages/PodList'
import DeploymentList from './pages/DeploymentList'
import StatefulSetList from './pages/StatefulSetList'
import DaemonSetList from './pages/DaemonSetList'
import ConfigMapList from './pages/ConfigMapList'
import SecretList from './pages/SecretList'
import PersistentVolumeList from './pages/PersistentVolumeList'
import PersistentVolumeClaimList from './pages/PersistentVolumeClaimList'
import StorageClassList from './pages/StorageClassList'
import HelmReleases from './pages/HelmReleases'
import HelmReleaseDetail from './pages/HelmReleaseDetail'
import HelmRepos from './pages/HelmRepos'

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Routes>
          <Route element={<MainLayout />}>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/namespaces" element={<NamespaceList />} />
            <Route path="/pods" element={<PodList />} />
            <Route path="/deployments" element={<DeploymentList />} />
            <Route path="/statefulsets" element={<StatefulSetList />} />
            <Route path="/daemonsets" element={<DaemonSetList />} />
            <Route path="/configmaps" element={<ConfigMapList />} />
            <Route path="/secrets" element={<SecretList />} />
            <Route path="/pvs" element={<PersistentVolumeList />} />
            <Route path="/pvcs" element={<PersistentVolumeClaimList />} />
            <Route path="/storageclasses" element={<StorageClassList />} />
            <Route path="/helm/releases" element={<HelmReleases />} />
            <Route path="/helm/releases/:namespace/:name" element={<HelmReleaseDetail />} />
            <Route path="/helm/repos" element={<HelmRepos />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App
