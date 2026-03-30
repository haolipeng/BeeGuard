import React, { useState, useEffect, useCallback } from 'react';
import { Table, Space, Button, Input, message } from 'antd';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import type { TableRowSelection } from 'antd/es/table/interface';

export interface PageTableProps<T> {
  columns: ColumnsType<T>;
  fetchData: (params: { page: number; limit: number; search?: string }) => Promise<any>;
  rowKey?: string | ((record: T) => string);
  searchPlaceholder?: string;
  extraActions?: React.ReactNode;
  batchActions?: (selectedKeys: React.Key[], clearSelection: () => void) => React.ReactNode;
  rowClassName?: (record: T) => string;
  pollInterval?: number;
}

function PageTable<T extends Record<string, any>>({
  columns,
  fetchData,
  rowKey = 'id',
  searchPlaceholder = '搜索...',
  extraActions,
  batchActions,
  rowClassName,
  pollInterval,
}: PageTableProps<T>) {
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [search, setSearch] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const result = await fetchData({ page, limit, search: search || undefined });
      // 兼容后端两种响应格式:
      // 格式1: { data: [...], pagination: { total_count: N } }
      // 格式2: { data: [...], total: N }
      // 格式3: { data: { list: [...], total: N } }
      let list: T[] = [];
      let count = 0;
      if (Array.isArray(result?.data)) {
        list = result.data;
        count = result.pagination?.total_count ?? result.total ?? list.length;
      } else if (result?.data?.list) {
        list = result.data.list;
        count = result.data.total ?? list.length;
      }
      setData(list);
      setTotal(count);
    } catch {
      message.error('数据加载失败');
    } finally {
      setLoading(false);
    }
  }, [fetchData, page, limit, search]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  useEffect(() => {
    if (!pollInterval) return;
    const timer = setInterval(loadData, pollInterval);
    return () => clearInterval(timer);
  }, [loadData, pollInterval]);

  const handleTableChange = (pagination: TablePaginationConfig) => {
    setPage(pagination.current || 1);
    setLimit(pagination.pageSize || 20);
  };

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const clearSelection = () => setSelectedRowKeys([]);

  const rowSelection: TableRowSelection<T> | undefined = batchActions
    ? {
        selectedRowKeys,
        onChange: setSelectedRowKeys,
      }
    : undefined;

  return (
    <div>
      <Space style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }} align="center">
        <Space>
          <Input.Search
            placeholder={searchPlaceholder}
            onSearch={handleSearch}
            allowClear
            style={{ width: 280 }}
            prefix={<SearchOutlined />}
          />
          {batchActions && selectedRowKeys.length > 0 && batchActions(selectedRowKeys, clearSelection)}
        </Space>
        <Space>
          {extraActions}
          <Button icon={<ReloadOutlined />} onClick={loadData}>
            刷新
          </Button>
        </Space>
      </Space>
      <Table<T>
        columns={columns}
        dataSource={data}
        rowKey={rowKey}
        loading={loading}
        rowSelection={rowSelection}
        rowClassName={rowClassName}
        pagination={{
          current: page,
          pageSize: limit,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t) => `共 ${t} 条`,
          pageSizeOptions: ['10', '20', '50', '100'],
        }}
        onChange={handleTableChange}
        scroll={{ x: 'max-content' }}
        size="middle"
      />
    </div>
  );
}

export default PageTable;
