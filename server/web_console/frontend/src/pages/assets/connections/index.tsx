import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '进程名', dataIndex: 'comm', key: 'comm', width: 120 },
  { title: '可执行路径', dataIndex: 'exe_path', key: 'exe_path', width: 200, ellipsis: true },
  { title: '协议', dataIndex: 'protocol', key: 'protocol', width: 70 },
  { title: '远端 IP', dataIndex: 'remote_ip', key: 'remote_ip', width: 140 },
  { title: '远端端口', dataIndex: 'remote_port', key: 'remote_port', width: 90 },
  { title: 'PID', dataIndex: 'pid', key: 'pid', width: 70 },
  { title: '事件时间', dataIndex: 'event_time', key: 'event_time', width: 170 },
];

const Connections: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('connection', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>网络连接</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索地址或进程..." />
    </div>
  );
};

export default Connections;
