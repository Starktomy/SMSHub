import {useState} from 'react';
import {useQuery, useMutation, useQueryClient} from '@tanstack/react-query';
import {devicesApi} from '@/api/devices';
import type {Device, CreateDeviceRequest} from '@/api/devices';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Switch} from '@/components/ui/switch';
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
import {toast} from 'sonner';
import {
    Plus,
    RefreshCw,
    Signal,
    Power,
    PowerOff,
    Plane,
    RotateCcw,
    Trash2,
    Edit,
    Wifi,
    WifiOff,
    Search,
} from 'lucide-react';

export default function Devices() {
    const queryClient = useQueryClient();
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
    const [editingDevice, setEditingDevice] = useState<Device | null>(null);
    const [formData, setFormData] = useState<CreateDeviceRequest>({
        name: '',
        serialPort: '',
        groupName: '',
        enabled: true,
    });

    // 获取设备列表
    const {data: devices, isLoading, refetch} = useQuery({
        queryKey: ['devices'],
        queryFn: devicesApi.list,
        refetchInterval: 10000,
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
    const {data: discoveredPorts} = useQuery({
        queryKey: ['discoveredPorts'],
        queryFn: devicesApi.discover,
    });

    // 创建设备
    const createMutation = useMutation({
        mutationFn: devicesApi.create,
        onSuccess: () => {
            toast.success('设备添加成功');
            setIsAddDialogOpen(false);
            setFormData({name: '', serialPort: '', groupName: '', enabled: true});
            queryClient.invalidateQueries({queryKey: ['devices']});
        },
        onError: (error: any) => {
            toast.error(error?.message || '添加设备失败');
        },
    });

    // 更新设备
    const updateMutation = useMutation({
        mutationFn: ({id, data}: {id: string; data: CreateDeviceRequest}) =>
            devicesApi.update(id, data),
        onSuccess: () => {
            toast.success('设备更新成功');
            setEditingDevice(null);
            queryClient.invalidateQueries({queryKey: ['devices']});
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
            queryClient.invalidateQueries({queryKey: ['devices']});
        },
        onError: () => {
            toast.error('删除设备失败');
        },
    });

    // 启用/禁用设备
    const toggleEnableMutation = useMutation({
        mutationFn: ({id, enabled}: {id: string; enabled: boolean}) =>
            enabled ? devicesApi.enable(id) : devicesApi.disable(id),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['devices']});
        },
        onError: () => {
            toast.error('操作失败');
        },
    });

    // 设置飞行模式
    const flymodeMutation = useMutation({
        mutationFn: ({id, enabled}: {id: string; enabled: boolean}) =>
            devicesApi.setFlymode(id, enabled),
        onSuccess: () => {
            toast.success('飞行模式设置成功');
            queryClient.invalidateQueries({queryKey: ['devices']});
        },
        onError: () => {
            toast.error('设置飞行模式失败');
        },
    });

    // 重启设备
    const rebootMutation = useMutation({
        mutationFn: devicesApi.reboot,
        onSuccess: () => {
            toast.success('重启命令已发送');
        },
        onError: () => {
            toast.error('重启失败');
        },
    });

    const handleSubmit = () => {
        if (!formData.serialPort) {
            toast.error('请选择串口');
            return;
        }
        if (editingDevice) {
            updateMutation.mutate({id: editingDevice.id, data: formData});
        } else {
            createMutation.mutate(formData);
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'online':
                return 'bg-green-500';
            case 'offline':
                return 'bg-gray-400';
            case 'error':
                return 'bg-red-500';
            default:
                return 'bg-gray-400';
        }
    };

    const getStatusText = (status: string) => {
        switch (status) {
            case 'online':
                return '在线';
            case 'offline':
                return '离线';
            case 'error':
                return '错误';
            default:
                return '未知';
        }
    };

    const getSignalBars = (level: number) => {
        const percentage = Math.round((level / 31) * 100);
        if (percentage >= 75) return 4;
        if (percentage >= 50) return 3;
        if (percentage >= 25) return 2;
        if (percentage > 0) return 1;
        return 0;
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-64">
                <RefreshCw className="w-8 h-8 animate-spin text-gray-400" />
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 头部 */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">设备管理</h1>
                    <p className="text-sm text-gray-500 mt-1">
                        管理多个 Air780 设备，支持批量操作
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
                                            setFormData({...formData, name: e.target.value})
                                        }
                                    />
                                </div>
                                <div>
                                    <label className="text-sm font-medium">串口</label>
                                    <Select
                                        value={formData.serialPort}
                                        onValueChange={(value) =>
                                            setFormData({...formData, serialPort: value})
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
                                            setFormData({...formData, groupName: e.target.value})
                                        }
                                    />
                                </div>
                                <div className="flex items-center justify-between">
                                    <label className="text-sm font-medium">立即启用</label>
                                    <Switch
                                        checked={formData.enabled}
                                        onCheckedChange={(checked) =>
                                            setFormData({...formData, enabled: checked})
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

            {/* 设备列表 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {devices?.map((device) => (
                    <Card key={device.id} className="relative overflow-hidden">
                        <CardHeader className="pb-2">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <div
                                        className={`w-3 h-3 rounded-full ${getStatusColor(
                                            device.status
                                        )} ${device.status === 'online' ? 'animate-pulse' : ''}`}
                                    />
                                    <CardTitle className="text-lg">
                                        {device.name || '未命名设备'}
                                    </CardTitle>
                                </div>
                                <span
                                    className={`text-xs px-2 py-1 rounded-full ${
                                        device.status === 'online'
                                            ? 'bg-green-100 text-green-700'
                                            : device.status === 'error'
                                            ? 'bg-red-100 text-red-700'
                                            : 'bg-gray-100 text-gray-700'
                                    }`}
                                >
                                    {getStatusText(device.status)}
                                </span>
                            </div>
                        </CardHeader>
                        <CardContent className="space-y-3">
                            {/* 设备信息 */}
                            <div className="grid grid-cols-2 gap-2 text-sm">
                                <div>
                                    <span className="text-gray-500">串口:</span>
                                    <span className="ml-1 font-mono text-xs">
                                        {device.serialPort?.split('/').pop()}
                                    </span>
                                </div>
                                <div>
                                    <span className="text-gray-500">运营商:</span>
                                    <span className="ml-1">{device.operator || '-'}</span>
                                </div>
                                <div>
                                    <span className="text-gray-500">号码:</span>
                                    <span className="ml-1 font-mono text-xs">
                                        {device.phoneNumber || '-'}
                                    </span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <span className="text-gray-500">信号:</span>
                                    <div className="flex items-end gap-0.5 h-4">
                                        {[1, 2, 3, 4].map((bar) => (
                                            <div
                                                key={bar}
                                                className={`w-1 rounded-sm transition-all ${
                                                    bar <= getSignalBars(device.signalLevel)
                                                        ? 'bg-green-500'
                                                        : 'bg-gray-200'
                                                }`}
                                                style={{height: `${bar * 4}px`}}
                                            />
                                        ))}
                                    </div>
                                </div>
                            </div>

                            {/* 分组 */}
                            {device.groupName && (
                                <div className="text-xs text-gray-500">
                                    分组: {device.groupName}
                                </div>
                            )}

                            {/* 飞行模式状态 */}
                            {device.flymode && (
                                <div className="flex items-center gap-1 text-xs text-orange-600">
                                    <Plane className="w-3 h-3" />
                                    飞行模式
                                </div>
                            )}

                            {/* 操作按钮 */}
                            <div className="flex items-center justify-between pt-2 border-t">
                                <div className="flex items-center gap-1">
                                    <Switch
                                        checked={device.enabled}
                                        onCheckedChange={(checked) =>
                                            toggleEnableMutation.mutate({
                                                id: device.id,
                                                enabled: checked,
                                            })
                                        }
                                    />
                                    <span className="text-xs text-gray-500">
                                        {device.enabled ? '已启用' : '已禁用'}
                                    </span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8"
                                        onClick={() =>
                                            flymodeMutation.mutate({
                                                id: device.id,
                                                enabled: !device.flymode,
                                            })
                                        }
                                        disabled={device.status !== 'online'}
                                        title={device.flymode ? '关闭飞行模式' : '开启飞行模式'}
                                    >
                                        <Plane
                                            className={`w-4 h-4 ${
                                                device.flymode ? 'text-orange-500' : ''
                                            }`}
                                        />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8"
                                        onClick={() => rebootMutation.mutate(device.id)}
                                        disabled={device.status !== 'online'}
                                        title="重启设备"
                                    >
                                        <RotateCcw className="w-4 h-4" />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8 text-red-500 hover:text-red-700"
                                        onClick={() => {
                                            if (confirm('确定要删除这个设备吗？')) {
                                                deleteMutation.mutate(device.id);
                                            }
                                        }}
                                        title="删除设备"
                                    >
                                        <Trash2 className="w-4 h-4" />
                                    </Button>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>

            {/* 空状态 */}
            {devices?.length === 0 && (
                <Card className="p-8 text-center">
                    <div className="text-gray-400 mb-4">
                        <Wifi className="w-16 h-16 mx-auto" />
                    </div>
                    <h3 className="text-lg font-medium text-gray-900 mb-2">
                        暂无设备
                    </h3>
                    <p className="text-gray-500 mb-4">
                        点击上方"添加设备"按钮添加您的第一个 Air780 设备
                    </p>
                    <Button onClick={() => setIsAddDialogOpen(true)}>
                        <Plus className="w-4 h-4 mr-2" />
                        添加设备
                    </Button>
                </Card>
            )}
        </div>
    );
}
