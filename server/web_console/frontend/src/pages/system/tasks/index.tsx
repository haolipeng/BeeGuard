import React, { useState, useCallback, useEffect } from 'react';
import { Typography, Button, Modal, Form, Input, Select, Space, Tag, message } from 'antd';
import { SendOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import PageTable from '../../../components/PageTable';
import { tasksApi, TaskSendPayload } from '../../../api/tasks';

const { Title, Text } = Typography;

const statusColors: Record<number, { color: string; text: string }> = {
  0: { color: 'blue', text: '已发送' },
  1: { color: 'processing', text: '执行中' },
  2: { color: 'green', text: '成功' },
  3: { color: 'red', text: '失败' },
  4: { color: 'default', text: '超时' },
};

const Tasks: React.FC = () => {
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();
  const [taskTypes, setTaskTypes] = useState<any[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    tasksApi.getTypes().then((res: any) => {
      setTaskTypes(res.data || res || []);
    }).catch(() => {});
  }, []);

  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await tasksApi.getHistory(params);
    return { data: res.data || [], total: res.total || 0 };
  }, []);

  const handleSend = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      const payload: TaskSendPayload = {
        agent_id: values.agent_id,
        task_type: values.task_type,
        task_name: values.task_name || '',
        parameters: values.parameters ? JSON.parse(values.parameters) : undefined,
      };
      await tasksApi.send(payload);
      message.success('任务已发送');
      setModalOpen(false);
      form.resetFields();
      setRefreshKey((k) => k + 1);
    } catch (err: any) {
      if (err?.errorFields) return;
      if (err instanceof SyntaxError) {
        message.error('参数 JSON 格式错误');
        return;
      }
      message.error('发送失败');
    } finally {
      setSubmitting(false);
    }
  };

  const columns: ColumnsType<any> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '任务 ID', dataIndex: 'task_id', key: 'task_id', width: 280, ellipsis: true },
    { title: 'Agent ID', dataIndex: 'agent_id', key: 'agent_id', width: 200, ellipsis: true },
    { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 140 },
    { title: '主机 IP', dataIndex: 'host_ip', key: 'host_ip', width: 120 },
    { title: '任务名', dataIndex: 'task_name', key: 'task_name', width: 140 },
    { title: '任务类型', dataIndex: 'task_type', key: 'task_type', width: 80 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => {
        const cfg = statusColors[status] || { color: 'default', text: '未知' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    { title: '结果', dataIndex: 'result_message', key: 'result_message', ellipsis: true },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 170 },
    { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
  ];

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>远程任务管理</Title>
      <PageTable
        key={refreshKey}
        columns={columns}
        fetchData={fetchData}
        searchPlaceholder="搜索 Agent ID 或主机名..."
        pollInterval={10000}
        extraActions={
          <Button type="primary" icon={<SendOutlined />} onClick={() => setModalOpen(true)}>
            发送任务
          </Button>
        }
      />
      <Modal
        title="发送远程任务"
        open={modalOpen}
        onOk={handleSend}
        onCancel={() => setModalOpen(false)}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="agent_id" label="目标 Agent ID" rules={[{ required: true, message: '请输入 Agent ID' }]}>
            <Input placeholder="例如：a1b2c3d4-xxxx-xxxx..." />
          </Form.Item>
          <Form.Item name="task_type" label="任务类型" rules={[{ required: true, message: '请选择任务类型' }]}>
            <Select placeholder="选择任务类型">
              {taskTypes.map((t: any) => (
                <Select.Option key={t.type || t.task_type} value={t.type || t.task_type}>
                  {t.name || t.task_name} ({t.plugin || t.plugin_name})
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="task_name" label="任务名称">
            <Input placeholder="可选，自定义任务名" />
          </Form.Item>
          <Form.Item name="parameters" label="参数 (JSON)">
            <Input.TextArea rows={3} placeholder='可选，例如：{"timeout": 60}' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Tasks;
