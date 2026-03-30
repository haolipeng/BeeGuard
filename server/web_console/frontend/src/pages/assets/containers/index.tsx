import React, { useCallback } from 'react';
import { Typography, Tag } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '容器 ID', dataIndex: 'container_id', key: 'container_id', width: 140, ellipsis: true },
  { title: '容器名', dataIndex: 'name', key: 'name', width: 160 },
  { title: '镜像', dataIndex: 'image_name', key: 'image_name', width: 200, ellipsis: true },
  { title: '运行时', dataIndex: 'runtime', key: 'runtime', width: 80 },
  {
    title: '状态',
    dataIndex: 'state',
    key: 'state',
    width: 100,
    render: (state: string) => (
      <Tag color={state === 'running' ? 'green' : 'default'}>{state}</Tag>
    ),
  },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Containers: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('container', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>容器列表</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索容器名或镜像..." />
    </div>
  );
};

export default Containers;
