import React, { useCallback } from 'react';
import { Typography, Progress } from 'antd';
import PageTable from '../../../components/PageTable';
import { baselineApi } from '../../../api/baseline';

const { Title } = Typography;

const columns = [
  { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 140 },
  { title: '主机名', dataIndex: 'hostname', key: 'hostname', width: 160 },
  { title: '检查模板数', dataIndex: 'template_count', key: 'template_count', width: 100 },
  { title: '通过数', dataIndex: 'pass_count', key: 'pass_count', width: 80 },
  { title: '失败数', dataIndex: 'fail_count', key: 'fail_count', width: 80 },
  {
    title: '合规率',
    dataIndex: 'compliance_rate',
    key: 'compliance_rate',
    width: 160,
    render: (rate: number) => (
      <Progress
        percent={Math.round((rate || 0) * 100)}
        size="small"
        strokeColor={rate >= 0.8 ? '#52c41a' : rate >= 0.5 ? '#faad14' : '#ff4d4f'}
      />
    ),
  },
  { title: '最后检查时间', dataIndex: 'last_checked_at', key: 'last_checked_at', width: 170 },
];

const BaselineHostView: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await baselineApi.getHostView(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>基线 - 主机视图</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索主机 IP..." />
    </div>
  );
};

export default BaselineHostView;
