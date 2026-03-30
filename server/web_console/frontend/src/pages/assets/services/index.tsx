import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '服务名', dataIndex: 'name', key: 'name', width: 200 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 90 },
  { title: '运行用户', dataIndex: 'run_user', key: 'run_user', width: 100 },
  { title: '路径', dataIndex: 'path', key: 'path', width: 240, ellipsis: true },
  { title: '描述', dataIndex: 'describe', key: 'describe', ellipsis: true },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Services: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('service', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>服务列表</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索服务名..." />
    </div>
  );
};

export default Services;
