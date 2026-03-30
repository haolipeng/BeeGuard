import React, { useState, useCallback, useEffect } from 'react';
import {
  Typography, Tabs, Table, Button, Space, Modal, Form, Input, Select, Switch,
  message, Popconfirm, Tag, Divider, Card,
} from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { whitelistApi, WhitelistRule, WhitelistConditions, WhitelistCondition } from '../../../api/whitelist';

const { Title, Text } = Typography;

const alertTypes = [
  { key: 'dangerous_command', label: '高危命令' },
  { key: 'reverse_shell', label: '反弹 Shell' },
  { key: 'privilege_escalation', label: '本地提权' },
  { key: 'abnormal_login', label: '异常登录' },
  { key: 'brute_force', label: '暴力破解' },
  { key: 'malicious_request', label: '恶意请求' },
  { key: 'network_attack', label: '网络攻击' },
  { key: 'malware_scan', label: '恶意文件' },
  { key: 'fileguard', label: '文件完整性' },
  { key: 'container_alert', label: '容器告警' },
];

const fieldOptions: Record<string, { label: string; value: string }[]> = {
  dangerous_command: [
    { label: '命令', value: 'command' }, { label: '用户', value: 'user' },
    { label: '主机IP', value: 'host_ip' }, { label: '进程名', value: 'process_name' },
  ],
  reverse_shell: [
    { label: '命令', value: 'command' }, { label: '目标地址', value: 'dst_addr' },
    { label: '主机IP', value: 'host_ip' },
  ],
  privilege_escalation: [
    { label: '命令', value: 'command' }, { label: '用户', value: 'user' },
    { label: '主机IP', value: 'host_ip' },
  ],
  abnormal_login: [
    { label: '源IP', value: 'source_ip' }, { label: '用户', value: 'user' },
    { label: '主机IP', value: 'host_ip' },
  ],
  brute_force: [
    { label: '源IP', value: 'source_ip' }, { label: '用户', value: 'user' },
    { label: '主机IP', value: 'host_ip' },
  ],
  malicious_request: [
    { label: 'URL', value: 'url' }, { label: '源IP', value: 'source_ip' },
    { label: '主机IP', value: 'host_ip' },
  ],
  network_attack: [
    { label: '源IP', value: 'source_ip' }, { label: '攻击类型', value: 'attack_type' },
    { label: '主机IP', value: 'host_ip' },
  ],
  malware_scan: [
    { label: '文件路径', value: 'file_path' }, { label: '恶意类型', value: 'malware_type' },
    { label: 'MD5', value: 'md5' }, { label: '主机IP', value: 'host_ip' },
  ],
  fileguard: [
    { label: '文件路径', value: 'file_path' }, { label: '用户', value: 'user' },
    { label: '主机IP', value: 'host_ip' },
  ],
  container_alert: [
    { label: '容器名', value: 'container_name' }, { label: '容器ID', value: 'container_id' },
    { label: '主机IP', value: 'host_ip' },
  ],
};

const operatorOptions = [
  { label: '等于', value: 'eq' },
  { label: '正则匹配', value: 'regex' },
  { label: '包含', value: 'contains' },
];

const Whitelist: React.FC = () => {
  const [activeType, setActiveType] = useState('dangerous_command');
  const [data, setData] = useState<WhitelistRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<WhitelistRule | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res: any = await whitelistApi.getList(activeType, { page, limit: 20 });
      setData(res.data || []);
      setTotal(res.total || 0);
    } catch {
      message.error('加载白名单失败');
    } finally {
      setLoading(false);
    }
  }, [activeType, page]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreate = () => {
    setEditingRule(null);
    form.resetFields();
    form.setFieldsValue({
      scope: 'global',
      conditions: { logic: 'and', rules: [{ field: '', operator: 'eq', value: '' }] },
    });
    setModalOpen(true);
  };

  const handleEdit = (record: WhitelistRule) => {
    setEditingRule(record);
    form.setFieldsValue({
      name: record.name,
      description: record.description,
      scope: record.scope,
      agent_ids: record.agent_ids,
      conditions: record.conditions,
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const payload: WhitelistRule = {
        name: values.name,
        description: values.description,
        scope: values.scope,
        agent_ids: values.scope === 'agent' ? values.agent_ids : undefined,
        conditions: values.conditions,
      };

      if (editingRule?.id) {
        await whitelistApi.update(activeType, editingRule.id, payload);
        message.success('规则更新成功');
      } else {
        await whitelistApi.create(activeType, payload);
        message.success('规则创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error('操作失败');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await whitelistApi.delete(activeType, id);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const handleToggle = async (id: number) => {
    try {
      await whitelistApi.toggle(activeType, id);
      message.success('状态切换成功');
      fetchData();
    } catch {
      message.error('操作失败');
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '规则名称', dataIndex: 'name', key: 'name', width: 180 },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    {
      title: '范围',
      dataIndex: 'scope',
      key: 'scope',
      width: 80,
      render: (scope: string) => (
        <Tag color={scope === 'global' ? 'blue' : 'cyan'}>{scope === 'global' ? '全局' : '指定'}</Tag>
      ),
    },
    {
      title: '条件',
      dataIndex: 'conditions',
      key: 'conditions',
      width: 200,
      ellipsis: true,
      render: (conditions: WhitelistConditions) => {
        if (!conditions?.rules) return '-';
        return conditions.rules.map((r, i) => (
          <span key={i}>
            {i > 0 && <Text type="secondary"> {conditions.logic} </Text>}
            <Text code>{r.field} {r.operator} {r.value}</Text>
          </span>
        ));
      },
    },
    { title: '命中数', dataIndex: 'hit_count', key: 'hit_count', width: 80 },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 70,
      render: (enabled: boolean, record: WhitelistRule) => (
        <Switch checked={enabled} size="small" onChange={() => handleToggle(record.id!)} />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: WhitelistRule) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          <Popconfirm title="确认删除此规则？" onConfirm={() => handleDelete(record.id!)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>白名单管理</Title>

      <Tabs
        activeKey={activeType}
        onChange={(key) => { setActiveType(key); setPage(1); }}
        items={alertTypes.map((t) => ({ key: t.key, label: t.label }))}
        style={{ marginBottom: 16 }}
      />

      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建规则
        </Button>
      </Space>

      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize: 20,
          total,
          onChange: setPage,
          showTotal: (t) => `共 ${t} 条`,
        }}
        size="middle"
        scroll={{ x: 'max-content' }}
      />

      <Modal
        title={editingRule ? '编辑规则' : '新建规则'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        width={700}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="规则名称" rules={[{ required: true, message: '请输入规则名称' }]}>
            <Input placeholder="例如：忽略 root 用户的 crontab 命令" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="规则描述（可选）" />
          </Form.Item>
          <Form.Item name="scope" label="作用范围" rules={[{ required: true }]}>
            <Select options={[{ label: '全局', value: 'global' }, { label: '指定 Agent', value: 'agent' }]} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.scope !== cur.scope}>
            {({ getFieldValue }) =>
              getFieldValue('scope') === 'agent' ? (
                <Form.Item name="agent_ids" label="Agent ID 列表" rules={[{ required: true, message: '请输入 Agent ID' }]}>
                  <Input.TextArea rows={2} placeholder="多个 ID 用逗号分隔" />
                </Form.Item>
              ) : null
            }
          </Form.Item>

          <Divider>匹配条件</Divider>

          <Form.Item name={['conditions', 'logic']} label="条件关系">
            <Select options={[{ label: 'AND（全部满足）', value: 'and' }, { label: 'OR（任一满足）', value: 'or' }]} />
          </Form.Item>

          <Form.List name={['conditions', 'rules']}>
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item name={[name, 'field']} rules={[{ required: true, message: '选择字段' }]}>
                      <Select
                        placeholder="字段"
                        style={{ width: 140 }}
                        options={fieldOptions[activeType] || []}
                      />
                    </Form.Item>
                    <Form.Item name={[name, 'operator']} rules={[{ required: true, message: '选择运算符' }]}>
                      <Select placeholder="运算符" style={{ width: 120 }} options={operatorOptions} />
                    </Form.Item>
                    <Form.Item name={[name, 'value']} rules={[{ required: true, message: '输入匹配值' }]}>
                      <Input placeholder="匹配值" style={{ width: 200 }} />
                    </Form.Item>
                    {fields.length > 1 && (
                      <MinusCircleOutlined onClick={() => remove(name)} style={{ color: '#ff4d4f' }} />
                    )}
                  </Space>
                ))}
                <Button type="dashed" onClick={() => add({ field: '', operator: 'eq', value: '' })} block icon={<PlusOutlined />}>
                  添加条件
                </Button>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>
    </div>
  );
};

export default Whitelist;
