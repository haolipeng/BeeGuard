import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '端口', dataIndex: 'port', key: 'port', width: 80 },
  { title: '协议', dataIndex: 'protocol', key: 'protocol', width: 80,
    render: (v: number) => v === 6 ? 'TCP' : v === 17 ? 'UDP' : String(v) },
  { title: '监听地址', dataIndex: 'listen_ip', key: 'listen_ip', width: 140 },
  { title: '监听进程', dataIndex: 'listen_process', key: 'listen_process', width: 140 },
  { title: '运行用户', dataIndex: 'run_user', key: 'run_user', width: 100 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Ports: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('port', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>端口列表</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索端口或进程..." />
    </div>
  );
};

export default Ports;
