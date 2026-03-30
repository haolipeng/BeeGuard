import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 130,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '进程名', dataIndex: 'name', key: 'name', width: 140 },
  { title: '进程路径', dataIndex: 'path', key: 'path', ellipsis: true },
  { title: '运行用户', dataIndex: 'run_name', key: 'run_name', width: 100 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '启动时间', dataIndex: 'start_time', key: 'start_time', width: 170 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Processes: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('process', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>进程列表</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索进程名..." />
    </div>
  );
};

export default Processes;
