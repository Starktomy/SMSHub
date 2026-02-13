import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { devicesApi } from '@/api/devices';
import type { Device, CreateDeviceRequest } from '@/api/devices';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { toast } from 'sonner';
import {
    Plus,
    RefreshCw,
    Plane,
    Trash2,
    Wifi,
    Search,
    Settings,
} from 'lucide-react';

import { SignalStrength } from '@/components/SignalStrength';
import { DeviceControlPanel } from '@/components/DeviceControlPanel';

// 状态颜色映射
const STATUS_COLORS = {
    online: 'bg-green-500',
    offline: 'bg-gray-400',
    error: 'bg-red-500',
};

export default function Devices() {
    const queryClient = useQueryClient();
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
    const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
    const [editingDevice, setEditingDevice] = useState<Device | null>(null);
    const [formData, setFormData] = useState<CreateDeviceRequest>({
        name: '',
        serialPort: '',
        groupName: '',
        enabled: true,
    });

    // 获取设备列表
    const { data: devices, isLoading, refetch } = useQuery({
        queryKey: ['devices'],
        queryFn: devicesApi.list,
        refetchInterval: 5000,
    });

    // 扫描串口
    const discoverMutation = useMutation({
        mutationFn: devicesApi.discover,
        onSuccess: (data) => {
            toast.success(`发现 ${data.ports?.length || 0} 个串口`);
        },
        onError: () => {
            toast.error('扫描串口失败');
        },
    });

    // 获取可用串口
    const { data: discoveredPorts } = useQuery({
        queryKey: ['discoveredPorts'],
        queryFn: devicesApi.discover,
    });

    // 创建设备
    const createMutation = useMutation({
        mutationFn: devicesApi.create,
        onSuccess: () => {
            toast.success('设备添加成功');
            setIsAddDialogOpen(false);
            setFormData({ name: '', serialPort: '', groupName: '', enabled: true });
            queryClient.invalidateQueries({ queryKey: ['devices'] });
        },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        onError: (error: any) => {
            toast.error(error?.message || '添加设备失败');
        },
    });

    // 更新设备
    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: CreateDeviceRequest }) =>
            devicesApi.update(id, data),
        onSuccess: () => {
            toast.success('设备更新成功');
            setEditingDevice(null);
            queryClient.invalidateQueries({ queryKey: ['devices'] });
        },
        onError: () => {
            toast.error('更新设备失败');
        },
    });

    // 删除设备
    const deleteMutation = useMutation({
        mutationFn: devicesApi.delete,
        onSuccess: () => {
            toast.success('设备删除成功');
            queryClient.invalidateQueries({ queryKey: ['devices'] });
        },
        onError: () => {
            toast.error('删除设备失败');
        },
    });

    const handleSubmit = () => {
        if (!formData.serialPort) {
            toast.error('请选择串口');
            return;
        }
        if (editingDevice) {
            updateMutation.mutate({ id: editingDevice.id, data: formData });
        } else {
            createMutation.mutate(formData);
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-64" data-testid="loading-spinner">
                <RefreshCw className="w-8 h-8 animate-spin text-gray-400" />
            </div>
        );
    }

    return (
        <div className="flex flex-col h-[calc(100vh-140px)]">
            {/* 头部：保持在顶端 */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6 shrink-0">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">设备管理</h1>
                    <p className="text-sm text-gray-500 mt-1">
                        管理多个 Air780 设备，支持实时控制
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => discoverMutation.mutate()}
                        disabled={discoverMutation.isPending}
                    >
                        <Search className="w-4 h-4 mr-2" />
                        扫描串口
                    </Button>
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => refetch()}
                    >
                        <RefreshCw className="w-4 h-4 mr-2" />
                        刷新
                    </Button>
                    <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
                        <DialogTrigger asChild>
                            <Button size="sm">
                                <Plus className="w-4 h-4 mr-2" />
                                添加设备
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>添加设备</DialogTitle>
                            </DialogHeader>
                            <div className="space-y-4 mt-4">
                                <div>
                                    <label className="text-sm font-medium">设备名称</label>
                                    <Input
                                        placeholder="如：香港卡1"
                                        value={formData.name}
                                        onChange={(e) =>
                                            setFormData({ ...formData, name: e.target.value })
                                        }
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">串口</label>
                                    <Select
                                        value={formData.serialPort}
                                        onValueChange={(value) =>
                                            setFormData({ ...formData, serialPort: value })
                                        }
                                    >
                                        <SelectTrigger>
                                            <SelectValue placeholder="选择串口" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {discoveredPorts?.ports?.map((port) => (
                                                <SelectItem key={port} value={port}>
                                                    {port}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div>
                                    <label className="text-sm font-medium">分组</label>
                                    <Input
                                        placeholder="可选，如：香港"
                                        value={formData.groupName}
                                        onChange={(e) =>
                                            setFormData({ ...formData, groupName: e.target.value })
                                        }
                                    />
                                </div>
                                <div className="flex items-center justify-between">
                                    <label className="text-sm font-medium">立即启用</label>
                                    <Switch
                                        checked={formData.enabled}
                                        onCheckedChange={(checked) =>
                                            setFormData({ ...formData, enabled: checked })
                                        }
                                    />
                                </div>
                                <Button
                                    className="w-full"
                                    onClick={handleSubmit}
                                    disabled={createMutation.isPending}
                                >
                                    {createMutation.isPending ? '添加中...' : '添加设备'}
                                </Button>
                            </div>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            {/* 主体：左右布局 */}
            <div className="flex-1 flex flex-col lg:flex-row gap-6 min-h-0 pb-10">
                {/* 左侧：精简设备列表 */}
                <Card className="w-full lg:w-80 flex flex-col shrink-0 overflow-hidden border-gray-200">
                    <CardHeader className="py-3 px-4 border-b bg-gray-50/50">
                        <CardTitle className="text-sm font-semibold flex items-center justify-between">
                            设备列表
                            <span className="text-xs font-normal text-gray-500 bg-gray-100 px-2 py-0.5 rounded-full">
                                {devices?.length || 0}
                            </span>
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1 overflow-y-auto p-2 space-y-1">
                        {devices?.map((device) => (
                            <div
                                key={device.id}
                                onClick={() => setSelectedDevice(device)}
                                className={`group relative flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-all border ${
                                    selectedDevice?.id === device.id
                                        ? 'bg-blue-50 border-blue-200 ring-1 ring-blue-100'
                                        : 'hover:bg-gray-50 border-transparent hover:border-gray-200'
                                }`}
                            >
                                {/* 状态指示灯 */}
                                <div className="relative shrink-0">
                                    <div
                                        className={`w-3 h-3 rounded-full ${
                                            STATUS_COLORS[device.status as keyof typeof STATUS_COLORS] || STATUS_COLORS.offline
                                        } ${device.status === 'online' ? 'animate-pulse shadow-[0_0_8px_rgba(34,197,94,0.4)]' : ''}`}
                                    />
                                </div>

                                {/* 设备基本信息 */}
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center justify-between">
                                        <p className={`text-sm font-medium truncate ${
                                            selectedDevice?.id === device.id ? 'text-blue-700' : 'text-gray-900'
                                        }`}>
                                            {device.name || '未命名设备'}
                                        </p>
                                        <SignalStrength level={device.signalLevel} />
                                    </div>
                                    <div className="flex items-center justify-between mt-0.5">
                                        <p className="text-xs text-gray-500 font-mono truncate">
                                            {device.serialPort?.split('/').pop()}
                                        </p>
                                        {device.flymode && (
                                            <Plane className="w-3 h-3 text-orange-500" />
                                        )}
                                    </div>
                                </div>

                                {/* 删除按钮（仅在非选中或悬停时显示） */}
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-7 w-7 opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500 transition-opacity shrink-0"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        if (confirm('确定要删除这个设备吗？')) {
                                            deleteMutation.mutate(device.id);
                                            if (selectedDevice?.id === device.id) setSelectedDevice(null);
                                        }
                                    }}
                                >
                                    <Trash2 className="w-3.5 h-3.5" />
                                </Button>

                                {/* 选中指示条 */}
                                {selectedDevice?.id === device.id && (
                                    <div className="absolute left-0 top-1/4 bottom-1/4 w-1 bg-blue-500 rounded-r-full" />
                                )}
                            </div>
                        ))}

                        {/* 空状态 */}
                        {devices?.length === 0 && (
                            <div className="py-12 text-center">
                                <Wifi className="w-12 h-12 mx-auto text-gray-200 mb-3" />
                                <p className="text-sm text-gray-500">暂无设备</p>
                            </div>
                        )}
                    </CardContent>
                </Card>

                {/* 右侧：沉浸式控制面板 */}
                <div className="flex-1 min-w-0">
                    {selectedDevice ? (
                        <div className="space-y-6 pb-12">
                            {/* 获取最新状态的设备对象 */}
                            <DeviceControlPanel
                                device={devices?.find(d => d.id === selectedDevice.id) || selectedDevice}
                            />
                        </div>
                    ) : (
                        <Card className="min-h-[400px] lg:h-full border-dashed border-2 bg-gray-50/50 flex flex-col items-center justify-center text-center p-12">
                            <div className="w-20 h-20 bg-white rounded-full shadow-sm flex items-center justify-center mb-6">
                                <Settings className="w-10 h-10 text-gray-300 animate-pulse" />
                            </div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-2">
                                欢迎进入设备管理中心
                            </h3>
                            <p className="text-gray-500 max-w-sm">
                                请在左侧列表中选择一台设备，以进行短信发送、飞行模式控制及模块重启等操作。
                            </p>
                        </Card>
                    )}
                </div>
            </div>
        </div>
    );
}
