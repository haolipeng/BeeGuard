import request, { PaginationParams } from './request';

export interface TaskSendPayload {
  agent_id: string;
  task_type: number;
  task_name: string;
  parameters?: Record<string, unknown>;
}

export const tasksApi = {
  send: (data: TaskSendPayload) =>
    request.post('/tasks/send', data),
  getHistory: (params?: PaginationParams & { agent_id?: string; status?: number; task_type?: number }) =>
    request.get('/tasks/history', { params }),
  getDetail: (id: number) =>
    request.get(`/tasks/history/${id}`),
  getTypes: () =>
    request.get('/tasks/types'),
};
