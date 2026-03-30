import React, { useState, useCallback } from 'react';
import { Typography, Row, Col, Card, Badge, Space, Descriptions, Spin } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  ClusterOutlined,
} from '@ant-design/icons';
import { usePolling } from '../../../hooks/usePolling';
import { statusApi } from '../../../api/system';

const { Title, Text } = Typography;

const ServiceStatus: React.FC = () => {
  const [serverStatus, setServerStatus] = useState<any>(null);
  const [dbStatus, setDbStatus] = useState<any>(null);
  const [agentsStatus, setAgentsStatus] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const fetchAll = useCallback(async () => {
    try {
      const [server, db, agents] = await Promise.allSettled([
        statusApi.getServerStatus(),
        statusApi.getDatabaseStatus(),
        statusApi.getAgentsStatus(),
      ]);
      if (server.status === 'fulfilled') setServerStatus(server.value);
      if (db.status === 'fulfilled') setDbStatus(db.value);
      if (agents.status === 'fulfilled') setAgentsStatus(agents.value);
    } finally {
      setLoading(false);
    }
  }, []);

  usePolling(fetchAll, 15000);

  const StatusIcon: React.FC<{ ok: boolean }> = ({ ok }) =>
    ok ? (
      <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 20 }} />
    ) : (
      <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 20 }} />
    );

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>服务状态</Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card
            title={
              <Space>
                <CloudServerOutlined />
                <Text style={{ color: '#fff' }}>Server 服务</Text>
              </Space>
            }
            style={{ border: '1px solid #303030' }}
            extra={<StatusIcon ok={serverStatus?.status === 'ok'} />}
          >
            <Descriptions column={1} size="small" labelStyle={{ color: 'rgba(255,255,255,0.45)' }} contentStyle={{ color: 'rgba(255,255,255,0.85)' }}>
              <Descriptions.Item label="状态">{serverStatus?.status || '未知'}</Descriptions.Item>
              <Descriptions.Item label="Goroutines">{serverStatus?.goroutines || '-'}</Descriptions.Item>
              <Descriptions.Item label="内存使用">{serverStatus?.mem_alloc || '-'}</Descriptions.Item>
              <Descriptions.Item label="系统内存">{serverStatus?.mem_sys || '-'}</Descriptions.Item>
              <Descriptions.Item label="GC 次数">{serverStatus?.num_gc || '-'}</Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card
            title={
              <Space>
                <DatabaseOutlined />
                <Text style={{ color: '#fff' }}>数据库</Text>
              </Space>
            }
            style={{ border: '1px solid #303030' }}
            extra={<StatusIcon ok={dbStatus?.status === 'ok'} />}
          >
            <Descriptions column={1} size="small" labelStyle={{ color: 'rgba(255,255,255,0.45)' }} contentStyle={{ color: 'rgba(255,255,255,0.85)' }}>
              <Descriptions.Item label="状态">{dbStatus?.status || '未知'}</Descriptions.Item>
              <Descriptions.Item label="活跃连接">{dbStatus?.in_use ?? '-'}</Descriptions.Item>
              <Descriptions.Item label="空闲连接">{dbStatus?.idle ?? '-'}</Descriptions.Item>
              <Descriptions.Item label="最大连接">{dbStatus?.max_open ?? '-'}</Descriptions.Item>
              <Descriptions.Item label="Ping 延迟">{dbStatus?.ping_ms ? `${dbStatus.ping_ms}ms` : '-'}</Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card
            title={
              <Space>
                <ClusterOutlined />
                <Text style={{ color: '#fff' }}>Agent 状态</Text>
              </Space>
            }
            style={{ border: '1px solid #303030' }}
            extra={<StatusIcon ok={(agentsStatus?.online || 0) > 0} />}
          >
            <Descriptions column={1} size="small" labelStyle={{ color: 'rgba(255,255,255,0.45)' }} contentStyle={{ color: 'rgba(255,255,255,0.85)' }}>
              <Descriptions.Item label="在线数">
                <Text style={{ color: '#52c41a' }}>{agentsStatus?.online || 0}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="离线数">
                <Text style={{ color: 'rgba(255,255,255,0.45)' }}>{agentsStatus?.offline || 0}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="总计">{agentsStatus?.total || 0}</Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ServiceStatus;
