import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '软件名称', dataIndex: 'name', key: 'name', width: 200 },
  { title: '版本', dataIndex: 'version', key: 'version', width: 140 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 80 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 160, ellipsis: true },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Software: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('software', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>软件列表</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索软件名..." />
    </div>
  );
};

export default Software;
