import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import App from './App';
import './styles/global.css';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

root.render(
  <React.StrictMode>
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: '#1668dc',
          colorBgContainer: '#1f1f1f',
          colorBgLayout: '#141414',
          borderRadius: 4,
        },
      }}
    >
      <BrowserRouter basename="/ui">
        <App />
      </BrowserRouter>
    </ConfigProvider>
  </React.StrictMode>
);
