import React from 'react';
import { Card, Statistic, Typography } from 'antd';
import type { StatisticProps } from 'antd';

const { Text } = Typography;

interface StatCardProps {
  title: string;
  value: number | string;
  suffix?: string;
  prefix?: React.ReactNode;
  valueStyle?: React.CSSProperties;
  description?: string;
  loading?: boolean;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  suffix,
  prefix,
  valueStyle,
  description,
  loading,
}) => {
  return (
    <Card
      size="small"
      style={{ borderRadius: 4, border: '1px solid #303030' }}
      loading={loading}
    >
      <Statistic
        title={<Text style={{ color: 'rgba(255,255,255,0.45)', fontSize: 13 }}>{title}</Text>}
        value={value}
        suffix={suffix}
        prefix={prefix}
        valueStyle={{ color: '#fff', fontSize: 28, ...valueStyle }}
      />
      {description && (
        <Text style={{ color: 'rgba(255,255,255,0.35)', fontSize: 12 }}>{description}</Text>
      )}
    </Card>
  );
};

export default StatCard;
