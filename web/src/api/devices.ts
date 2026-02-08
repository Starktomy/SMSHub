import apiClient from './client';

export interface Device {
  id: string;
  name: string;
  serialPort: string;
  status: 'online' | 'offline' | 'error';
  phoneNumber: string;
  operator: string;
  signalLevel: number;
  flymode: boolean;
  enabled: boolean;
  groupName: string;
  lastSeenAt: number;
  createdAt: number;
  updatedAt: number;
}

export interface CreateDeviceRequest {
  name: string;
  serialPort: string;
  groupName?: string;
  enabled?: boolean;
}

export interface UpdateDeviceRequest {
  name: string;
  serialPort: string;
  groupName?: string;
  enabled?: boolean;
}

export interface BatchSendRequest {
  recipients: string[];
  content: string;
  deviceId?: string;
  strategy?: 'auto' | 'round_robin' | 'random' | 'signal_best';
}

export interface BatchSendResult {
  recipient: string;
  messageId: string;
  deviceId: string;
  success: boolean;
  error?: string;
}

export interface DiscoverResponse {
  ports: string[];
}

export interface BatchSendResponse {
  results: BatchSendResult[];
}

export const devicesApi = {
  // Device CRUD
  list: () => apiClient.get<Device[]>('/devices'),
  get: (id: string) => apiClient.get<Device>(`/devices/${id}`),
  create: (data: CreateDeviceRequest) => apiClient.post<Device>('/devices', data),
  update: (id: string, data: UpdateDeviceRequest) => apiClient.put(`/devices/${id}`, data),
  delete: (id: string) => apiClient.delete(`/devices/${id}`),

  // Device actions
  enable: (id: string) => apiClient.post(`/devices/${id}/enable`),
  disable: (id: string) => apiClient.post(`/devices/${id}/disable`),
  setFlymode: (id: string, enabled: boolean) => apiClient.post(`/devices/${id}/flymode`, { enabled }),
  reboot: (id: string) => apiClient.post(`/devices/${id}/reboot`),
  getStatus: (id: string) => apiClient.get(`/devices/${id}/status`),

  // Discovery and groups
  discover: () => apiClient.get<DiscoverResponse>('/devices/discover'),
  getGroups: () => apiClient.get<{ groups: string[] }>('/devices/groups'),
  getStats: () => apiClient.get<Record<string, number>>('/devices/stats'),

  // SMS
  sendSMS: (id: string, to: string, content: string) =>
    apiClient.post(`/devices/${id}/sms`, { to, content }),
  autoSendSMS: (to: string, content: string, strategy?: string) =>
    apiClient.post('/sms/send', { to, content, strategy }),
  batchSend: (data: BatchSendRequest) =>
    apiClient.post<BatchSendResponse>('/sms/batch', data),
};
