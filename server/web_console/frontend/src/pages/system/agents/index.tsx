import React, { useCallback } from 'react';
import { Typography, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import PageTable from '../../../components/PageTable';
import { systemApi } from '../../../api/system';

const { Title } = Typography;

const columns: ColumnsType<any> = [
  { title: 'Agent ID', dataIndex: 'agent_id', key: 'agent_id', width: 280, ellipsis: true },
  { title: '主机名', dataIndex: 'hostname', key: 'hostname', width: 160 },
  { title: 'IP 地址', dataIndex: 'ipv4', key: 'ipv4', width: 140 },
  { title: '操作系统', dataIndex: 'os', key: 'os', width: 140 },
  { title: '版本', dataIndex: 'agent_version', key: 'agent_version', width: 100 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    render: (status: number) => (
      <Tag color={status === 1 ? 'green' : 'default'}>{status === 1 ? '在线' : '离线'}</Tag>
    ),
  },
  { title: '最后心跳', dataIndex: 'last_heartbeat', key: 'last_heartbeat', width: 170 },
  { title: '注册时间', dataIndex: 'created_at', key: 'created_at', width: 170 },
];

const Agents: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await systemApi.getAgents(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>Agent ��理</Title>
      <PageTable
        columns={columns}
        fetchData={fetchData}
        rowKey="agent_id"
        searchPlaceholder="搜索主机名或 IP..."
        pollInterval={30000}
      />
    </div>
  );
};

export default Agents;
