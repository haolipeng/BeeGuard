import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '用户名', dataIndex: 'name', key: 'name', width: 120 },
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 80 },
  { title: '权限', dataIndex: 'permission', key: 'permission', width: 100 },
  { title: '登录 Shell', dataIndex: 'login_type', key: 'login_type', width: 180 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Accounts: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('account', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>账户列表</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索用户名..." />
    </div>
  );
};

export default Accounts;
