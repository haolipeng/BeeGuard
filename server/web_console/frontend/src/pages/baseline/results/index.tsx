import React, { useCallback } from 'react';
import { Typography, Tag } from 'antd';
import PageTable from '../../../components/PageTable';
import { baselineApi } from '../../../api/baseline';

const { Title } = Typography;

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 140 },
  { title: '模板名', dataIndex: 'template_name', key: 'template_name', width: 160 },
  {
    title: '结果',
    dataIndex: 'pass',
    key: 'pass',
    width: 80,
    render: (pass: boolean) => <Tag color={pass ? 'green' : 'red'}>{pass ? '通过' : '不通过'}</Tag>,
  },
  { title: '通过项', dataIndex: 'pass_count', key: 'pass_count', width: 80 },
  { title: '失败项', dataIndex: 'fail_count', key: 'fail_count', width: 80 },
  { title: '检查时间', dataIndex: 'checked_at', key: 'checked_at', width: 170 },
];

const BaselineResults: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await baselineApi.getResults(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>检��结果</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索主机 IP..." />
    </div>
  );
};

export default BaselineResults;
