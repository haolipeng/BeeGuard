import React, { useState, useCallback } from 'react';
import { Row, Col, Card, Typography, Badge, Space } from 'antd';
import {
  DesktopOutlined,
  AlertOutlined,
  BugOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import StatCard from '../../components/StatCard';
import { usePolling } from '../../hooks/usePolling';
import { dashboardApi } from '../../api/dashboard';

const { Title, Text } = Typography;

const Dashboard: React.FC = () => {
  const [hostStatus, setHostStatus] = useState<any>(null);
  const [threatCount, setThreatCount] = useState<any>(null);
  const [hourlyStats, setHourlyStats] = useState<any[]>([]);
  const [vulnTop5, setVulnTop5] = useState<any[]>([]);
  const [baselineTop5, setBaselineTop5] = useState<any[]>([]);
  const [serviceStatus, setServiceStatus] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const fetchAll = useCallback(async () => {
    try {
      const [host, threat, hourly, vuln, baseline, service] = await Promise.allSettled([
        dashboardApi.getHostStatusSummary(),
        dashboardApi.getThreatTypeTotalCount(),
        dashboardApi.getAlertHourlyStats(),
        dashboardApi.getHostVulnTop5(),
        dashboardApi.getHostBaselineFailTop5(),
        dashboardApi.getServiceStatusOverview(),
      ]);
      if (host.status === 'fulfilled') setHostStatus(host.value?.data || host.value);
      if (threat.status === 'fulfilled') setThreatCount(threat.value?.data || threat.value);
      if (hourly.status === 'fulfilled') setHourlyStats(hourly.value?.data || []);
      if (vuln.status === 'fulfilled') setVulnTop5(vuln.value?.data || []);
      if (baseline.status === 'fulfilled') setBaselineTop5(baseline.value?.data || []);
      if (service.status === 'fulfilled') setServiceStatus(service.value?.data || service.value);
    } finally {
      setLoading(false);
    }
  }, []);

  usePolling(fetchAll);

  const renderStatusBadge = (status: boolean | undefined, label: string) => (
    <Space>
      <Badge status={status ? 'success' : 'error'} />
      <Text style={{ color: 'rgba(255,255,255,0.85)' }}>{label}</Text>
      <Text style={{ color: status ? '#52c41a' : '#ff4d4f', fontSize: 12 }}>
        {status ? '正常' : '异常'}
      </Text>
    </Space>
  );

  return (
    <div>
      <Title level={4} style={{ color: '#fff', marginBottom: 20 }}>安全概览</Title>

      <Row gutter={[16, 16]}>
        {/* 在线主机统计 */}
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="在线主机"
            value={hostStatus?.online || 0}
            suffix={`/ ${hostStatus?.total || 0}`}
            prefix={<DesktopOutlined />}
            valueStyle={{ color: '#52c41a' }}
            description={`离线 ${hostStatus?.offline || 0} 台`}
            loading={loading}
          />
        </Col>

        {/* 待处理告警 */}
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="待处理告警"
            value={threatCount?.total || 0}
            prefix={<AlertOutlined />}
            valueStyle={{ color: threatCount?.total > 0 ? '#ff4d4f' : '#52c41a' }}
            description="需要立即处理"
            loading={loading}
          />
        </Col>

        {/* 主机漏洞 */}
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="主机漏洞"
            value={vulnTop5?.length || 0}
            prefix={<BugOutlined />}
            valueStyle={{ color: '#faad14' }}
            description="存在漏洞的主机"
            loading={loading}
          />
        </Col>

        {/* 基线不合规 */}
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="基线不合规"
            value={baselineTop5?.length || 0}
            prefix={<SafetyCertificateOutlined />}
            valueStyle={{ color: '#faad14' }}
            description="存在不合规项的主机"
            loading={loading}
          />
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        {/* 告警趋势 */}
        <Col xs={24} lg={16}>
          <Card
            title={<Text style={{ color: '#fff' }}>告警趋势 (24h)</Text>}
            style={{ border: '1px solid #303030' }}
            bodyStyle={{ minHeight: 240 }}
          >
            {hourlyStats.length > 0 ? (
              <div style={{ color: 'rgba(255,255,255,0.45)' }}>
                {hourlyStats.map((item: any, idx: number) => (
                  <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0', borderBottom: '1px solid #303030' }}>
                    <Text style={{ color: 'rgba(255,255,255,0.65)' }}>{item.hour || item.time}</Text>
                    <Text style={{ color: '#1668dc' }}>{item.count || 0}</Text>
                  </div>
                ))}
              </div>
            ) : (
              <div style={{ textAlign: 'center', padding: 40, color: 'rgba(255,255,255,0.25)' }}>暂无数据</div>
            )}
          </Card>
        </Col>

        {/* 服务状态 */}
        <Col xs={24} lg={8}>
          <Card
            title={<Text style={{ color: '#fff' }}>服务状态</Text>}
            style={{ border: '1px solid #303030' }}
            bodyStyle={{ minHeight: 240 }}
          >
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              {renderStatusBadge(serviceStatus?.server?.status === 'ok', 'Server 服务')}
              {renderStatusBadge(serviceStatus?.database?.status === 'ok', '数据库')}
              {renderStatusBadge(serviceStatus?.agents?.online > 0, 'Agent 在线')}
            </Space>
            <div style={{ marginTop: 24 }}>
              <Text style={{ color: 'rgba(255,255,255,0.45)', fontSize: 12 }}>
                Agent 在线：{serviceStatus?.agents?.online || 0} / {serviceStatus?.agents?.total || 0}
              </Text>
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        {/* 漏洞 TOP5 */}
        <Col xs={24} lg={12}>
          <Card
            title={<Text style={{ color: '#fff' }}>主机漏洞 TOP5</Text>}
            style={{ border: '1px solid #303030' }}
          >
            {vulnTop5.length > 0 ? vulnTop5.slice(0, 5).map((item: any, idx: number) => (
              <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', borderBottom: '1px solid #303030' }}>
                <Text style={{ color: 'rgba(255,255,255,0.65)' }} ellipsis>{item.host_ip || item.hostname}</Text>
                <Text style={{ color: '#faad14' }}>{item.count || 0}</Text>
              </div>
            )) : (
              <div style={{ textAlign: 'center', padding: 20, color: 'rgba(255,255,255,0.25)' }}>暂无数据</div>
            )}
          </Card>
        </Col>

        {/* 基线不合规 TOP5 */}
        <Col xs={24} lg={12}>
          <Card
            title={<Text style={{ color: '#fff' }}>基线不合规 TOP5</Text>}
            style={{ border: '1px solid #303030' }}
          >
            {baselineTop5.length > 0 ? baselineTop5.slice(0, 5).map((item: any, idx: number) => (
              <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', borderBottom: '1px solid #303030' }}>
                <Text style={{ color: 'rgba(255,255,255,0.65)' }} ellipsis>{item.host_ip || item.hostname}</Text>
                <Text style={{ color: '#ff4d4f' }}>{item.fail_count || 0}</Text>
              </div>
            )) : (
              <div style={{ textAlign: 'center', padding: 20, color: 'rgba(255,255,255,0.25)' }}>暂无数据</div>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
