import React, { useState, useCallback } from 'react';
import { Typography, Button, Space, Modal, Form, Input, message, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import PageTable from '../../../components/PageTable';
import { systemApi } from '../../../api/system';

const { Title } = Typography;

const Users: React.FC = () => {
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();
  const [refreshKey, setRefreshKey] = useState(0);

  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await systemApi.getUsers(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      await systemApi.createUser(values);
      message.success('用户创建成功');
      setModalOpen(false);
      form.resetFields();
      setRefreshKey((k) => k + 1);
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error('创建失败');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await systemApi.deleteUser(id);
      message.success('删除成功');
      setRefreshKey((k) => k + 1);
    } catch {
      message.error('删除失败');
    }
  };

  const columns: ColumnsType<any> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username', key: 'username', width: 180 },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 170 },
    { title: '最后登录', dataIndex: 'last_login', key: 'last_login', width: 170 },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: any, record: any) => (
        <Popconfirm title="确认删除此用户？" onConfirm={() => handleDelete(record.id)}>
          <Button type="link" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>用户管理</Title>
      <PageTable
        key={refreshKey}
        columns={columns}
        fetchData={fetchData}
        searchPlaceholder="搜索用户名..."
        extraActions={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            新建用户
          </Button>
        }
      />
      <Modal title="新建用户" open={modalOpen} onOk={handleCreate} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }, { min: 6, message: '至少6位' }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Users;
