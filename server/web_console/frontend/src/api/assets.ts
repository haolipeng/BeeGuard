import request, { PaginationParams } from './request';

// 后端各资产类型有独立路径（同时支持单数和复数 key）
const assetRouteMap: Record<string, string> = {
  host:        '/assets/host/hosts',
  hosts:       '/assets/host/hosts',
  port:        '/assets/port/ports',
  ports:       '/assets/port/ports',
  account:     '/assets/account/accounts',
  accounts:    '/assets/account/accounts',
  process:     '/assets/process/processes',
  processes:   '/assets/process/processes',
  database:    '/assets/database/databases',
  databases:   '/assets/database/databases',
  web:         '/assets/web/webs',
  webs:        '/assets/web/webs',
  service:     '/assets/system/systems',
  services:    '/assets/system/systems',
  container:   '/assets/container/containers',
  containers:  '/assets/container/containers',
  image:       '/assets/container/images',
  images:      '/assets/container/images',
  kmod:        '/assets/kmod/list',
  kmods:       '/assets/kmod/list',
  env:         '/assets/env/list',
  envs:        '/assets/env/list',
  software:    '/assets/software/list',
  connection:  '/assets/connect/list',
  connections: '/assets/connect/list',
};

export const assetsApi = {
  getHosts: (params?: PaginationParams & { search?: string }) =>
    request.get('/assets/host/hosts', { params }),
  getFingerprint: (type: string, params?: PaginationParams & { host_id?: string; search?: string }) =>
    request.get(assetRouteMap[type] || `/assets/host/hosts`, { params }),
  getOsTypeStats: () =>
    request.get('/assets/view/os-type-stats'),
  getHostStats: () =>
    request.get('/assets/view/host-stats'),
  getContainerStats: () =>
    request.get('/assets/view/container-stats'),
  getAccountStats: () =>
    request.get('/assets/view/account-stats'),
  getLatestAssetsTop5: () =>
    request.get('/assets/view/latest-assets-top5'),
};
