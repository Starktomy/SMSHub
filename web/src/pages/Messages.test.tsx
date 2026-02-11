import { render, screen, fireEvent, waitFor, waitForElementToBeRemoved } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import Messages from '../pages/Messages';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as messagesApi from '../api/messages';
import { devicesApi } from '@/api/devices';

// Mock dependencies - use relative path as component uses relative path
vi.mock('../api/messages', () => ({
  getConversations: vi.fn(),
  getConversationMessages: vi.fn(),
  deleteConversation: vi.fn(),
  deleteMessage: vi.fn(),
  clearMessages: vi.fn(),
}));

// Mock devices API using alias as component might use alias for devices
vi.mock('@/api/devices', () => ({
  devicesApi: {
    list: vi.fn(),
    sendSMS: vi.fn(),
    autoSendSMS: vi.fn(),
  },
}));

vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    warning: vi.fn(),
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

describe('Messages Page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders conversations list', async () => {
    const mockConversations = [
      {
        peer: '13800138000',
        messageCount: 5,
        lastMessage: {
          id: '1',
          content: 'Hello',
          createdAt: Date.now(),
          type: 'incoming',
        },
      },
    ];

    (messagesApi.getConversations as any).mockResolvedValue(mockConversations);
    (messagesApi.getConversationMessages as any).mockResolvedValue([]);
    (devicesApi.list as any).mockResolvedValue([]);

    render(
      <QueryClientProvider client={createTestQueryClient()}>
        <Messages />
      </QueryClientProvider>
    );

    // Wait for loading spinner to be removed
    await waitForElementToBeRemoved(() => screen.queryByTestId('loading-spinner'));

    // Peer number appears in list and header (when selected)
    const peerElements = screen.getAllByText('13800138000');
    expect(peerElements.length).toBeGreaterThan(0);
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });

  it('sends message when clicking send button', async () => {
    const mockConversations = [
      {
        peer: '13800138000',
        messageCount: 1,
        lastMessage: {
          id: '1',
          content: 'Hello',
          createdAt: Date.now(),
          type: 'incoming',
        },
      },
    ];

    (messagesApi.getConversations as any).mockResolvedValue(mockConversations);
    (messagesApi.getConversationMessages as any).mockResolvedValue([]);
    (devicesApi.list as any).mockResolvedValue([]);
    (devicesApi.autoSendSMS as any).mockResolvedValue({});

    render(
      <QueryClientProvider client={createTestQueryClient()}>
        <Messages />
      </QueryClientProvider>
    );

    // Wait for loading spinner to be removed
    await waitForElementToBeRemoved(() => screen.queryByTestId('loading-spinner'));

    // Peer number appears in list and header
    const peerElements = screen.getAllByText('13800138000');
    expect(peerElements.length).toBeGreaterThan(0);

    // Click the conversation to ensure it's selected (though useEffect should do it)
    // We click the first occurrence (usually the list item)
    fireEvent.click(peerElements[0]);

    // Type message
    const input = screen.getByPlaceholderText('输入消息内容...');
    fireEvent.change(input, { target: { value: 'Test Reply' } });

    // Click send
    const sendButton = screen.getByText('发送');
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(devicesApi.autoSendSMS).toHaveBeenCalledWith('13800138000', 'Test Reply');
    });
  });
});
