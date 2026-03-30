import React from 'react';
import { Tag } from 'antd';

interface StatusTagProps {
  status: number;
}

const statusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'red', text: '待处理' },
  1: { color: 'green', text: '已处理' },
  2: { color: 'default', text: '已忽略' },
};

const StatusTag: React.FC<StatusTagProps> = ({ status }) => {
  const config = statusMap[status] || { color: 'default', text: '未知' };
  return <Tag color={config.color}>{config.text}</Tag>;
};

export default StatusTag;
