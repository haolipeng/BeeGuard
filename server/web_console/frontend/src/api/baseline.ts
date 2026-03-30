import request, { PaginationParams } from './request';

export const baselineApi = {
  getTemplates: (params?: PaginationParams) =>
    request.get('/baseline/templates', { params }),
  getTemplate: (id: number) =>
    request.get(`/baseline/templates/${id}`),
  getResults: (params?: PaginationParams & { template_id?: number; host_id?: string }) =>
    request.get('/baseline/details', { params }),
  getHostView: (params?: PaginationParams) =>
    request.get('/baseline/host_views/host', { params }),
  getItemView: (params?: PaginationParams & { template_id?: number }) =>
    request.get('/baseline/item_views/item', { params }),
  getCardStatistics: (params?: PaginationParams) =>
    request.get('/baseline/card_statistics', { params }),
};
