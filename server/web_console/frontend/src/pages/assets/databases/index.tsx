import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 140 },
  { title: '数据库类型', dataIndex: 'db_type', key: 'db_type', width: 120 },
  { title: '版本', dataIndex: 'version', key: 'version', width: 140 },
  { title: '端口', dataIndex: 'port', key: 'port', width: 80 },
  { title: '绑定地址', dataIndex: 'bind_addr', key: 'bind_addr', width: 140 },
  { title: '数据目录', dataIndex: 'data_dir', key: 'data_dir', width: 220 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Databases: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('database', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>数据库实例</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索数据库类型..." />
    </div>
  );
};

export default Databases;
