import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Spin } from 'antd';
import MainLayout from './layouts/MainLayout';

const Login = lazy(() => import('./pages/login'));
const Dashboard = lazy(() => import('./pages/dashboard'));
const Hosts = lazy(() => import('./pages/assets/hosts'));
const Ports = lazy(() => import('./pages/assets/ports'));
const Processes = lazy(() => import('./pages/assets/processes'));
const Accounts = lazy(() => import('./pages/assets/accounts'));
const Software = lazy(() => import('./pages/assets/software'));
const Services = lazy(() => import('./pages/assets/services'));
const Containers = lazy(() => import('./pages/assets/containers'));
const Images = lazy(() => import('./pages/assets/images'));
const Databases = lazy(() => import('./pages/assets/databases'));
const WebServices = lazy(() => import('./pages/assets/webservices'));
const KMod = lazy(() => import('./pages/assets/kmod'));
const Envs = lazy(() => import('./pages/assets/envs'));
const Connections = lazy(() => import('./pages/assets/connections'));
const Alerts = lazy(() => import('./pages/alerts'));
const Whitelist = lazy(() => import('./pages/alerts/whitelist'));
const BaselineTemplates = lazy(() => import('./pages/baseline/templates'));
const BaselineResults = lazy(() => import('./pages/baseline/results'));
const BaselineHostView = lazy(() => import('./pages/baseline/host-view'));
const BaselineItemView = lazy(() => import('./pages/baseline/item-view'));
const Users = lazy(() => import('./pages/system/users'));
const Agents = lazy(() => import('./pages/system/agents'));
const Tasks = lazy(() => import('./pages/system/tasks'));
const ServiceStatus = lazy(() => import('./pages/system/service-status'));

const Loading: React.FC = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
    <Spin size="large" />
  </div>
);

const App: React.FC = () => {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          {/* 资产中心 */}
          <Route path="assets/hosts" element={<Hosts />} />
          <Route path="assets/ports" element={<Ports />} />
          <Route path="assets/processes" element={<Processes />} />
          <Route path="assets/accounts" element={<Accounts />} />
          <Route path="assets/software" element={<Software />} />
          <Route path="assets/services" element={<Services />} />
          <Route path="assets/containers" element={<Containers />} />
          <Route path="assets/images" element={<Images />} />
          <Route path="assets/databases" element={<Databases />} />
          <Route path="assets/webservices" element={<WebServices />} />
          <Route path="assets/kmod" element={<KMod />} />
          <Route path="assets/envs" element={<Envs />} />
          <Route path="assets/connections" element={<Connections />} />
          {/* 入侵检测 */}
          <Route path="alerts" element={<Alerts />} />
          <Route path="alerts/whitelist" element={<Whitelist />} />
          {/* 基线合规 */}
          <Route path="baseline/templates" element={<BaselineTemplates />} />
          <Route path="baseline/results" element={<BaselineResults />} />
          <Route path="baseline/host-view" element={<BaselineHostView />} />
          <Route path="baseline/item-view" element={<BaselineItemView />} />
          {/* 系统管理 */}
          <Route path="system/users" element={<Users />} />
          <Route path="system/agents" element={<Agents />} />
          <Route path="system/tasks" element={<Tasks />} />
          <Route path="system/service-status" element={<ServiceStatus />} />
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Suspense>
  );
};

export default App;
