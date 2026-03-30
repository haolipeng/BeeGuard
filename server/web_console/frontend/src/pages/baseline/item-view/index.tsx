import React, { useCallback } from 'react';
import { Typography, Tag } from 'antd';
import PageTable from '../../../components/PageTable';
import { baselineApi } from '../../../api/baseline';

const { Title } = Typography;

const columns = [
  { title: '检查项', dataIndex: 'item_name', key: 'item_name', width: 240 },
  { title: '所属模板', dataIndex: 'template_name', key: 'template_name', width: 160 },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
  {
    title: '等级',
    dataIndex: 'level',
    key: 'level',
    width: 80,
    render: (level: string) => {
      const colorMap: Record<string, string> = { high: 'red', medium: 'orange', low: 'blue' };
      return <Tag color={colorMap[level] || 'default'}>{level}</Tag>;
    },
  },
  { title: '通过主机数', dataIndex: 'pass_host_count', key: 'pass_host_count', width: 100 },
  { title: '失败主机数', dataIndex: 'fail_host_count', key: 'fail_host_count', width: 100 },
];

const BaselineItemView: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await baselineApi.getItemView(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>基线 - 检查项视图</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索检查项..." />
    </div>
  );
};

export default BaselineItemView;
