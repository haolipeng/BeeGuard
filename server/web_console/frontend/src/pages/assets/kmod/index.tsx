import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '模块名', dataIndex: 'name', key: 'name', width: 200 },
  { title: '大小', dataIndex: 'size', key: 'size', width: 100 },
  { title: '引用计数', dataIndex: 'refcount', key: 'refcount', width: 90 },
  { title: '被引用', dataIndex: 'used_by', key: 'used_by', width: 160, ellipsis: true },
  { title: '状态', dataIndex: 'state', key: 'state', width: 80 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const KMod: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('kmod', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>内核模块</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索模块名..." />
    </div>
  );
};

export default KMod;
