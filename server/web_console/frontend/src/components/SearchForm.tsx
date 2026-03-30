import React from 'react';
import { Form, Input, Select, Button, Space, DatePicker } from 'antd';
import { SearchOutlined, ClearOutlined } from '@ant-design/icons';

export interface SearchField {
  name: string;
  label: string;
  type: 'input' | 'select' | 'dateRange';
  options?: { label: string; value: string | number }[];
  placeholder?: string;
}

interface SearchFormProps {
  fields: SearchField[];
  onSearch: (values: Record<string, any>) => void;
  onReset?: () => void;
}

const SearchForm: React.FC<SearchFormProps> = ({ fields, onSearch, onReset }) => {
  const [form] = Form.useForm();

  const handleReset = () => {
    form.resetFields();
    onReset?.();
  };

  return (
    <Form form={form} layout="inline" onFinish={onSearch} style={{ marginBottom: 16 }}>
      {fields.map((field) => (
        <Form.Item key={field.name} name={field.name} label={field.label}>
          {field.type === 'input' && (
            <Input placeholder={field.placeholder || `请输入${field.label}`} allowClear style={{ width: 180 }} />
          )}
          {field.type === 'select' && (
            <Select
              placeholder={field.placeholder || `请选择${field.label}`}
              allowClear
              style={{ width: 140 }}
              options={field.options}
            />
          )}
          {field.type === 'dateRange' && (
            <DatePicker.RangePicker style={{ width: 240 }} />
          )}
        </Form.Item>
      ))}
      <Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
            搜索
          </Button>
          <Button onClick={handleReset} icon={<ClearOutlined />}>
            重置
          </Button>
        </Space>
      </Form.Item>
    </Form>
  );
};

export default SearchForm;
