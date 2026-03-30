import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, message } from 'antd';
import { UserOutlined, LockOutlined, CloudServerOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../../api/system';

const { Title, Text } = Typography;

const Login: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const res: any = await authApi.login(values.username, values.password);
      // 后端返回格式: { code: 200, data: { token: "eyJ...", user_info: {...} } }
      // 经过 axios 拦截器 response.data 后, res 即为上述 JSON body
      const token = res?.data?.token;
      if (!token) {
        message.error('登录失败：服务器未返回有效 token');
        return;
      }
      localStorage.setItem('token', token);
      message.success('登录成功');
      navigate('/dashboard', { replace: true });
    } catch {
      message.error('用户名或密码错误');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      height: '100vh',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      background: '#141414',
    }}>
      <Card style={{ width: 400, background: '#1f1f1f', border: '1px solid #303030' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <CloudServerOutlined style={{ fontSize: 48, color: '#1668dc' }} />
          <Title level={3} style={{ color: '#fff', marginTop: 12, marginBottom: 4 }}>
            BeeGuard
          </Title>
          <Text style={{ color: 'rgba(255,255,255,0.45)' }}>安全管理平台</Text>
        </div>
        <Form onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Login;
