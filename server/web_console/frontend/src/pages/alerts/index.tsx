import React, { useState, useCallback } from 'react';
import { Typography, Tabs, Button, Space, Select, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import PageTable from '../../components/PageTable';
import StatusTag from '../../components/StatusTag';
import { alertsApi, AlertType } from '../../api/alerts';

const { Title, Text } = Typography;

const alertTabs: { key: AlertType; label: string }[] = [
  { key: 'dangerous_command', label: '高危命令' },
  { key: 'reverse_shell', label: '反弹 Shell' },
  { key: 'privilege_escalation', label: '本地提权' },
  { key: 'abnormal_login', label: '异常登录' },
  { key: 'brute_force', label: '暴力破解' },
  { key: 'malicious_request', label: '恶意请求' },
  { key: 'network_attack', label: '网络攻击' },
  { key: 'malware_scan', label: '恶意文件' },
  { key: 'fileguard', label: '文件完整性' },
];

const baseColumns = (type: AlertType): ColumnsType<any> => [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  {
    title: '主机 IP',
    dataIndex: type === 'reverse_shell' ? 'victim_ip' : 'host_ip',
    key: 'host_ip',
    width: 130,
    render: (ip: string) => ip?.split(',')[0] || ip,
  },
  { title: '主机名', dataIndex: 'host_name', key: 'host_name', width: 120 },
];

const alertTypeColumns: Partial<Record<AlertType, ColumnsType<any>>> = {
  dangerous_command: [
    { title: '命令', dataIndex: 'command', key: 'command', ellipsis: true },
    { title: '用户', dataIndex: 'user', key: 'user', width: 80 },
    { title: '命令类型', dataIndex: 'command_type', key: 'command_type', width: 140 },
  ],
  reverse_shell: [
    { title: '命令', dataIndex: 'command_line', key: 'command_line', ellipsis: true },
    { title: 'Shell 类型', dataIndex: 'shell_type', key: 'shell_type', width: 100 },
    { title: '目标地址', dataIndex: 'target_host', key: 'target_host', width: 130 },
    { title: '目标端口', dataIndex: 'target_port', key: 'target_port', width: 80 },
  ],
  privilege_escalation: [
    { title: '进程路径', dataIndex: 'process_path', key: 'process_path', ellipsis: true },
    { title: '提权用户', dataIndex: 'escalated_user', key: 'escalated_user', width: 100 },
    { title: '父进程', dataIndex: 'parent_process', key: 'parent_process', width: 140, ellipsis: true },
  ],
  abnormal_login: [
    { title: '源 IP', dataIndex: 'source_ip', key: 'source_ip', width: 130 },
    { title: '登录用户', dataIndex: 'login_user', key: 'login_user', width: 100 },
    { title: '风险等级', dataIndex: 'risk_level', key: 'risk_level', width: 80 },
    { title: '登录时间', dataIndex: 'login_time', key: 'login_time', width: 170 },
  ],
  brute_force: [
    { title: '源 IP', dataIndex: 'source_ip', key: 'source_ip', width: 130 },
    { title: '攻击类型', dataIndex: 'attack_type', key: 'attack_type', width: 80 },
    { title: '用户名', dataIndex: 'username', key: 'username', width: 100 },
    { title: '尝试次数', dataIndex: 'attempt_count', key: 'attempt_count', width: 80 },
  ],
  malicious_request: [
    { title: '策略名称', dataIndex: 'policy_name', key: 'policy_name', width: 140 },
    { title: '策略类型', dataIndex: 'policy_type', key: 'policy_type', width: 100 },
    { title: '恶意域名', dataIndex: 'malicious_domain', key: 'malicious_domain', width: 180, ellipsis: true },
    { title: '请求次数', dataIndex: 'request_count', key: 'request_count', width: 80 },
  ],
  network_attack: [
    { title: '攻击者 IP', dataIndex: 'attacker_ip', key: 'attacker_ip', width: 130 },
    { title: '目标端口', dataIndex: 'target_port', key: 'target_port', width: 80 },
    { title: '漏洞名称', dataIndex: 'vulnerability_name', key: 'vulnerability_name', width: 180, ellipsis: true },
    { title: '攻击次数', dataIndex: 'attack_count', key: 'attack_count', width: 80 },
  ],
  malware_scan: [
    { title: '文件名', dataIndex: 'file_name', key: 'file_name', width: 140 },
    { title: '文件路径', dataIndex: 'file_path', key: 'file_path', ellipsis: true },
    { title: '威胁类型', dataIndex: 'threat_type', key: 'threat_type', width: 100 },
    { title: '恶意家族', dataIndex: 'malware_family', key: 'malware_family', width: 140 },
    { title: 'MD5', dataIndex: 'file_md5', key: 'file_md5', width: 280, ellipsis: true },
  ],
  fileguard: [
    { title: '文件路径', dataIndex: 'file_path', key: 'file_path', ellipsis: true },
    { title: '操作', dataIndex: 'threat_action', key: 'threat_action', width: 80 },
    { title: '规则名称', dataIndex: 'rule_name', key: 'rule_name', width: 140 },
    { title: '操作用户', dataIndex: 'operator_user', key: 'operator_user', width: 100 },
  ],
  container_alert: [
    { title: '容器 ID', dataIndex: 'container_id', key: 'container_id', width: 140, ellipsis: true },
    { title: '容器名', dataIndex: 'container_name', key: 'container_name', width: 140 },
    { title: '告警详情', dataIndex: 'detail', key: 'detail', ellipsis: true },
  ],
};

const Alerts: React.FC = () => {
  const [activeTab, setActiveTab] = useState<AlertType>('dangerous_command');
  const [statusFilter, setStatusFilter] = useState<number | undefined>(undefined);

  const fetchData = useCallback(async (params: { page: number; limit: number; search?: string }) => {
    const res: any = await alertsApi.getList(activeTab, {
      ...params,
      status: statusFilter,
    });
    return { data: res.data || [], total: res.total || 0 };
  }, [activeTab, statusFilter]);

  const handleBatchAction = async (ids: React.Key[], status: number, clearSelection: () => void) => {
    try {
      await alertsApi.batchUpdateStatus(activeTab, ids as number[], status);
      message.success('批量操作成功');
      clearSelection();
    } catch {
      message.error('操作失败');
    }
  };

  const tailColumns: ColumnsType<any> = [
    { title: '告警时间', dataIndex: 'created_at', key: 'created_at', width: 170 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: number) => <StatusTag status={status} />,
    },
  ];

  const typeSpecificCols = alertTypeColumns[activeTab] || [];
  const columns = [...baseColumns(activeTab), ...typeSpecificCols, ...tailColumns];

  const batchActions = (selectedKeys: React.Key[], clearSelection: () => void) => (
    <Space>
      <Text style={{ color: 'rgba(255,255,255,0.65)' }}>已选 {selectedKeys.length} 条</Text>
      <Button size="small" onClick={() => handleBatchAction(selectedKeys, 1, clearSelection)}>
        批量已处理
      </Button>
      <Button size="small" onClick={() => handleBatchAction(selectedKeys, 2, clearSelection)}>
        批量忽略
      </Button>
    </Space>
  );

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>告警列表</Title>

      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as AlertType)}
        items={alertTabs.map((tab) => ({ key: tab.key, label: tab.label }))}
        style={{ marginBottom: 16 }}
      />

      <Space style={{ marginBottom: 12 }}>
        <Select
          placeholder="状态筛选"
          allowClear
          style={{ width: 140 }}
          value={statusFilter}
          onChange={(v) => setStatusFilter(v)}
          options={[
            { label: '全部', value: undefined },
            { label: '待处理', value: 0 },
            { label: '已处理', value: 1 },
            { label: '已忽略', value: 2 },
          ]}
        />
      </Space>

      <PageTable
        key={`${activeTab}-${statusFilter}`}
        columns={columns}
        fetchData={fetchData}
        searchPlaceholder="搜索 IP、主机名..."
        batchActions={batchActions}
        rowClassName={(record: any) => record.whitelist_hit ? 'whitelist-hit-row' : ''}
        pollInterval={30000}
      />
    </div>
  );
};

export default Alerts;
