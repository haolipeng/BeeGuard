import request, { PaginationParams } from './request';

export interface WhitelistCondition {
  field: string;
  operator: 'eq' | 'regex' | 'contains';
  value: string;
}

export interface WhitelistConditions {
  logic: 'and' | 'or';
  rules: WhitelistCondition[];
}

export interface WhitelistRule {
  id?: number;
  name: string;
  description?: string;
  scope: 'global' | 'agent';
  agent_ids?: string;
  conditions: WhitelistConditions;
  enabled?: boolean;
  hit_count?: number;
  created_by?: string;
  created_at?: string;
  updated_at?: string;
}

export const whitelistApi = {
  getTypes: () => request.get('/whitelist/types'),
  getList: (alertType: string, params?: PaginationParams) =>
    request.get(`/whitelist/${alertType}`, { params }),
  getDetail: (alertType: string, id: number) =>
    request.get(`/whitelist/${alertType}/${id}`),
  create: (alertType: string, data: WhitelistRule) =>
    request.post(`/whitelist/${alertType}`, data),
  update: (alertType: string, id: number, data: Partial<WhitelistRule>) =>
    request.put(`/whitelist/${alertType}/${id}`, data),
  delete: (alertType: string, id: number) =>
    request.delete(`/whitelist/${alertType}/${id}`),
  toggle: (alertType: string, id: number) =>
    request.post(`/whitelist/${alertType}/${id}/toggle`),
};
