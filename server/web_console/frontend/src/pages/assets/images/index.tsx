import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '镜像 ID', dataIndex: 'image_id', key: 'image_id', width: 140, ellipsis: true },
  { title: '镜像名', dataIndex: 'image_name', key: 'image_name', width: 200, ellipsis: true },
  { title: '版本', dataIndex: 'image_version', key: 'image_version', width: 120 },
  { title: '运行时', dataIndex: 'runtime', key: 'runtime', width: 80 },
  { title: '容器数', dataIndex: 'container_count', key: 'container_count', width: 80 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Images: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('image', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>镜像列表</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索镜像名..." />
    </div>
  );
};

export default Images;
