import React, { useCallback } from 'react';
import { Typography, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns: ColumnsType<any> = [
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 160 },
  { title: 'IP 地址', dataIndex: 'host_ip', key: 'host_ip', width: 160,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '操作系统', dataIndex: 'os_type', key: 'os_type', width: 120 },
  { title: '系统版本', dataIndex: 'os_version', key: 'os_version', width: 120 },
  { title: 'Agent 版本', dataIndex: 'agent_version', key: 'agent_version', width: 130 },
  {
    title: '状态',
    dataIndex: 'agent_status',
    key: 'agent_status',
    width: 80,
    render: (status: number) => (
      <Tag color={status === 1 ? 'green' : 'default'}>{status === 1 ? '在线' : '离线'}</Tag>
    ),
  },
  { title: '最后心跳', dataIndex: 'last_heartbeat', key: 'last_heartbeat', width: 170 },
];

const Hosts: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getHosts(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>主机列表</Title>
      <PageTable
        columns={columns}
        fetchData={fetchData}
        searchPlaceholder="搜索主机名或 IP..."
        pollInterval={30000}
      />
    </div>
  );
};

export default Hosts;
