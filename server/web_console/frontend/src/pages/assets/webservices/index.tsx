import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 140 },
  { title: '服务名', dataIndex: 'name', key: 'name', width: 160 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 100 },
  { title: '端口', dataIndex: 'port', key: 'port', width: 80 },
  { title: '域名', dataIndex: 'domain', key: 'domain', width: 200 },
  { title: '路径', dataIndex: 'path', key: 'path', width: 200 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const WebServices: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('web', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>Web 服务</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索服务名或域名..." />
    </div>
  );
};

export default WebServices;
