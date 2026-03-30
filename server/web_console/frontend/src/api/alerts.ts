import request, { PaginationParams } from './request';

export type AlertType =
  | 'dangerous_command'
  | 'reverse_shell'
  | 'privilege_escalation'
  | 'abnormal_login'
  | 'brute_force'
  | 'malicious_request'
  | 'network_attack'
  | 'malware_scan'
  | 'fileguard'
  | 'container_alert';

export interface AlertQueryParams extends PaginationParams {
  status?: number;
  host_ip?: string;
  search?: string;
  whitelist_hit?: boolean;
}

// 后端每种告警类型有独立的路由路径
const alertRouteMap: Record<AlertType, { list: string; status: string }> = {
  dangerous_command:   { list: '/alerts/command/commands',    status: '/alerts/command/commands/status' },
  reverse_shell:       { list: '/alerts/shell/shells',       status: '/alerts/shell/shells/status' },
  privilege_escalation:{ list: '/alerts/local/alerts',       status: '/alerts/local/alerts/status' },
  abnormal_login:      { list: '/alerts/login/alerts',       status: '/alerts/login/alerts/status' },
  brute_force:         { list: '/alerts/passwd/alerts',      status: '/alerts/passwd/alerts/status' },
  malicious_request:   { list: '/alerts/request/alerts',     status: '/alerts/request/alerts/status' },
  network_attack:      { list: '/alerts/network/attacks',    status: '/alerts/network/attacks/status' },
  malware_scan:        { list: '/alerts/file/scans',         status: '/alerts/file/scans/status' },
  fileguard:           { list: '/alerts/fileguard/alerts',   status: '/alerts/fileguard/alerts/status' },
  container_alert:     { list: '/alerts/command/commands',    status: '/alerts/command/commands/status' },
};

export const alertsApi = {
  getList: (type: AlertType, params?: AlertQueryParams) =>
    request.get(alertRouteMap[type].list, { params }),
  getDetail: (type: AlertType, id: number) =>
    request.get(`${alertRouteMap[type].list}/${id}`),
  updateStatus: (type: AlertType, id: number, status: number) =>
    request.post(`${alertRouteMap[type].status}/${id}`, { status }),
  batchUpdateStatus: (type: AlertType, ids: number[], status: number) =>
    Promise.all(ids.map(id => request.post(`${alertRouteMap[type].status}/${id}`, { status }))),
};
