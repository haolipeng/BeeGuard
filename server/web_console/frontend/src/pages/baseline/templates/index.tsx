import React, { useCallback } from 'react';
import { Typography } from 'antd';
import PageTable from '../../../components/PageTable';
import { baselineApi } from '../../../api/baseline';

const { Title } = Typography;

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  { title: '模板名称', dataIndex: 'name', key: 'name', width: 200 },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
  { title: '检查项数', dataIndex: 'item_count', key: 'item_count', width: 80 },
  { title: '适用系统', dataIndex: 'os_type', key: 'os_type', width: 120 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 170 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
];

const BaselineTemplates: React.FC = () => {
  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await baselineApi.getTemplates(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>基线模板</Title>
      <PageTable columns={columns} fetchData={fetchData} searchPlaceholder="搜索模板名..." />
    </div>
  );
};

export default BaselineTemplates;
