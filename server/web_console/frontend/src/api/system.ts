import request from './request';

export const authApi = {
  login: (username: string, password: string) =>
    request.post('/auth/login', { username, password }),
  logout: () =>
    request.post('/auth/logout'),
  getUserInfo: () =>
    request.get('/user/info'),
};

export const systemApi = {
  getUsers: (params?: { page?: number; limit?: number }) =>
    request.get('/system/users', { params }),
  createUser: (data: { username: string; password: string }) =>
    request.post('/system/users/create', data),
  updateUser: (id: number, data: { password?: string }) =>
    request.get(`/system/users/edit/${id}`, { params: data }),
  deleteUser: (id: number) =>
    request.post(`/system/users/delete/${id}`),
  getAgents: (params?: { page?: number; limit?: number; search?: string }) =>
    request.get('/system/agents', { params }),
};

export const statusApi = {
  getServerStatus: () => request.get('/status/server'),
  getDatabaseStatus: () => request.get('/status/database'),
  getAgentsStatus: () => request.get('/status/agents'),
  getOverview: () =>
    Promise.allSettled([
      request.get('/status/server'),
      request.get('/status/database'),
      request.get('/status/agents'),
    ]).then(([server, database, agents]) => ({
      data: {
        server: server.status === 'fulfilled' ? server.value?.data || server.value : null,
        database: database.status === 'fulfilled' ? database.value?.data || database.value : null,
        agents: agents.status === 'fulfilled' ? agents.value?.data || agents.value : null,
      },
    })),
};
