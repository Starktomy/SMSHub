import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import Devices from '../pages/Devices';
import { devicesApi } from '@/api/devices';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Mock dependencies with the exact path used in the component
vi.mock('@/api/devices', () => ({
  devicesApi: {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    discover: vi.fn(),
    enable: vi.fn(),
    disable: vi.fn(),
    setFlymode: vi.fn(),
    reboot: vi.fn(),
  },
}));

// Mock toast
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Setup QueryClient
const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

describe('Devices Page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders device list correctly', async () => {
    const mockDevices = [
      {
        id: '1',
        name: 'Test Device',
        serialPort: '/dev/ttyUSB0',
        status: 'online',
        phoneNumber: '13800138000',
        operator: 'China Mobile',
        signalLevel: 20,
        flymode: false,
        enabled: true,
      },
    ];

    (devicesApi.list as Mock).mockResolvedValue(mockDevices);
    (devicesApi.discover as Mock).mockResolvedValue({ ports: [] });

    render(
      <QueryClientProvider client={createTestQueryClient()}>
        <Devices />
      </QueryClientProvider>
    );

    // Wait for data to be loaded
    await waitFor(() => {
        expect(screen.getByText('Test Device')).toBeInTheDocument();
    });

    expect(screen.getByText('设备管理')).toBeInTheDocument();
    expect(screen.getByText('Test Device')).toBeInTheDocument();
    expect(screen.getByText('13800138000')).toBeInTheDocument();
  });

  it('opens add device dialog when clicking add button', async () => {
    (devicesApi.list as Mock).mockResolvedValue([]);
    (devicesApi.discover as Mock).mockResolvedValue({ ports: ['/dev/ttyUSB0'] });

    render(
      <QueryClientProvider client={createTestQueryClient()}>
        <Devices />
      </QueryClientProvider>
    );

    // Wait for rendering
    await waitFor(() => {
        expect(screen.getByText('设备管理')).toBeInTheDocument();
    });

    // There might be multiple "添加设备" buttons (header and empty state), click the first one
    const addButtons = screen.getAllByText('添加设备');
    fireEvent.click(addButtons[0]);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText('设备名称')).toBeInTheDocument();
    });
  });
});
