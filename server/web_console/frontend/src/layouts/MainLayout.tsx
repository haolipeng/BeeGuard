import React, { useState, useEffect } from 'react';
import { Layout, Menu, Dropdown, Button, Space, Typography } from 'antd';
import {
  DashboardOutlined,
  DesktopOutlined,
  AlertOutlined,
  SafetyCertificateOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
  ClusterOutlined,
  CloudServerOutlined,
  SendOutlined,
  HeartOutlined,
  UnorderedListOutlined,
  StopOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { authApi } from '../api/system';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

interface UserInfo {
  username: string;
}

const menuItems = [
  {
    key: 'dashboard',
    icon: <DashboardOutlined />,
    label: '安全概览',
  },
  {
    key: 'assets',
    icon: <DesktopOutlined />,
    label: '资产中心',
    children: [
      { key: 'assets/hosts', label: '主机列表' },
      { key: 'assets/ports', label: '端口' },
      { key: 'assets/processes', label: '进程' },
      { key: 'assets/accounts', label: '账户' },
      { key: 'assets/software', label: '软件' },
      { key: 'assets/services', label: '服务' },
      { key: 'assets/containers', label: '容器' },
      { key: 'assets/images', label: '镜像' },
      { key: 'assets/databases', label: '数据库' },
      { key: 'assets/webservices', label: 'Web 服务' },
      { key: 'assets/kmod', label: '内核模块' },
      { key: 'assets/envs', label: '环境变量' },
      { key: 'assets/connections', label: '网络连接' },
    ],
  },
  {
    key: 'alerts-group',
    icon: <AlertOutlined />,
    label: '入侵检测',
    children: [
      { key: 'alerts', icon: <UnorderedListOutlined />, label: '告警列表' },
      { key: 'alerts/whitelist', icon: <StopOutlined />, label: '白名单' },
    ],
  },
  {
    key: 'baseline',
    icon: <SafetyCertificateOutlined />,
    label: '基线合规',
    children: [
      { key: 'baseline/templates', label: '基线模板' },
      { key: 'baseline/results', label: '检查结果' },
      { key: 'baseline/host-view', label: '主机视图' },
      { key: 'baseline/item-view', label: '检查项视图' },
    ],
  },
  {
    key: 'system',
    icon: <SettingOutlined />,
    label: '系统管理',
    children: [
      { key: 'system/users', icon: <UserOutlined />, label: '用户管理' },
      { key: 'system/agents', icon: <ClusterOutlined />, label: 'Agent 管理' },
      { key: 'system/tasks', icon: <SendOutlined />, label: '远程任务' },
      { key: 'system/service-status', icon: <HeartOutlined />, label: '服务状态' },
    ],
  },
];

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);

  const currentPath = location.pathname.replace('/ui/', '').replace('/ui', '');

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token || token === 'undefined' || token === 'null') {
      localStorage.removeItem('token');
      navigate('/login', { replace: true });
      return;
    }
    authApi.getUserInfo().then((res: any) => {
      // 后端返回: { code: 200, data: { id, username, name, role } }
      setUserInfo(res?.data || res);
    }).catch(() => {
      localStorage.removeItem('token');
      navigate('/login', { replace: true });
    });
  }, [navigate]);

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(`/${key}`);
  };

  const handleLogout = async () => {
    try {
      await authApi.logout();
    } finally {
      localStorage.removeItem('token');
      navigate('/login', { replace: true });
    }
  };

  const selectedKeys = [currentPath || 'dashboard'];
  const openKeys = menuItems
    .filter((item) => item.children?.some((child) => currentPath.startsWith(child.key)))
    .map((item) => item.key);

  const userMenu = {
    items: [
      {
        key: 'logout',
        icon: <LogoutOutlined />,
        label: '退出登录',
        onClick: handleLogout,
      },
    ],
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        width={220}
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          borderRight: '1px solid #303030',
        }}
      >
        <div style={{
          height: 48,
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'flex-start',
          padding: collapsed ? 0 : '0 16px',
          borderBottom: '1px solid #303030',
        }}>
          <CloudServerOutlined style={{ fontSize: 24, color: '#1668dc' }} />
          {!collapsed && (
            <Text strong style={{ color: '#fff', fontSize: 16, marginLeft: 10 }}>
              BeeGuard
            </Text>
          )}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          defaultOpenKeys={openKeys}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <Layout style={{ marginLeft: collapsed ? 80 : 220, transition: 'margin-left 0.2s' }}>
        <Header style={{
          background: '#1f1f1f',
          padding: '0 24px',
          display: 'flex',
          justifyContent: 'flex-end',
          alignItems: 'center',
          borderBottom: '1px solid #303030',
          height: 48,
        }}>
          <Dropdown menu={userMenu} placement="bottomRight">
            <Button type="text" style={{ color: 'rgba(255,255,255,0.85)' }}>
              <Space>
                <UserOutlined />
                {userInfo?.username || '用户'}
              </Space>
            </Button>
          </Dropdown>
        </Header>
        <Content style={{ margin: 16, minHeight: 'calc(100vh - 80px)' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
