import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { assetsApi } from '../../../api/assets';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 150,
    render: (ip: string) => ip?.split(',')[0] || ip },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
  { title: '变量名', dataIndex: 'var_name', key: 'var_name', width: 200 },
  { title: '变量值', dataIndex: 'var_value', key: 'var_value', ellipsis: true },
  { title: '可疑原因', dataIndex: 'suspicious_reasons', key: 'suspicious_reasons', width: 200, ellipsis: true },
  { title: '来源', dataIndex: 'source', key: 'source', width: 100 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const Envs: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await assetsApi.getFingerprint('env', params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>可疑环境变量</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索变量名..." />
    </div>
  );
};

export default Envs;
