import { useState, useEffect } from 'react';
import { Activity, RotateCcw, Send, Signal, Wifi, Edit2, Check, X } from 'lucide-react';
import { toast } from 'sonner';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { devicesApi } from '@/api/devices';
import type { Device } from '@/api/devices';
import { SignalStrength } from '@/components/SignalStrength';

interface DeviceControlPanelProps {
    device: Device;
}

// 状态颜色和文本映射
const STATUS_CONFIG = {
    online: {
        text: '在线',
        color: 'text-green-600',
        bgColor: 'bg-green-100',
        borderColor: 'border-green-300',
        dotColor: 'bg-green-600',
    },
    offline: {
        text: '离线',
        color: 'text-red-600',
        bgColor: 'bg-red-100',
        borderColor: 'border-red-300',
        dotColor: 'bg-red-600',
    },
};

// 飞行模式状态配置
const FLYMODE_CONFIG = {
    true: {
        text: '已启用',
        color: 'text-orange-600',
        bgColor: 'bg-orange-100',
        borderColor: 'border-orange-300',
    },
    false: {
        text: '已禁用',
        color: 'text-green-600',
        bgColor: 'bg-green-100',
        borderColor: 'border-green-300',
    },
};

export function DeviceControlPanel({ device }: DeviceControlPanelProps) {
    const queryClient = useQueryClient();
    const [to, setTo] = useState('');
    const [content, setContent] = useState('');

    // 手机号编辑状态
    const [isEditingPhone, setIsEditingPhone] = useState(false);
    const [tempPhone, setTempPhone] = useState('');

    useEffect(() => {
        if (device) {
            setTempPhone(device.phoneNumber || '');
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [device.phoneNumber]);

    // 更新设备 Mutation
    const updateDeviceMutation = useMutation({
        mutationFn: (data: { id: string; phoneNumber: string }) => {
            return devicesApi.update(data.id, {
                name: device.name,
                serialPort: device.serialPort,
                groupName: device.groupName,
                enabled: device.enabled,
                phoneNumber: data.phoneNumber,
            });
        },
        onSuccess: () => {
            toast.success('手机号已更新');
            setIsEditingPhone(false);
            queryClient.invalidateQueries({ queryKey: ['devices'] });
        },
        onError: (error) => {
            console.error('更新失败:', error);
            toast.error('更新失败');
        },
    });

    // 发送短信 Mutation
    const sendSMSMutation = useMutation({
        mutationFn: (data: { to: string; content: string }) => {
            return devicesApi.sendSMS(device.id, data.to, data.content);
        },
        onSuccess: () => {
            toast.success('短信下发成功，等待确认...');
            setTo('');
            setContent('');
        },
        onError: (error) => {
            console.error('发送失败:', error);
            toast.error('发送失败');
        },
    });

    // 设置飞行模式 Mutation
    const setFlymodeMutation = useMutation({
        mutationFn: (enabled: boolean) => {
            return devicesApi.setFlymode(device.id, enabled);
        },
        onSuccess: () => {
            toast.success('设置成功');
            queryClient.invalidateQueries({ queryKey: ['devices'] });
        },
        onError: (error) => {
            console.error('操作失败:', error);
            toast.error('操作失败');
        },
    });

    // 重启模块 Mutation
    const rebootMcuMutation = useMutation({
        mutationFn: () => {
            return devicesApi.reboot(device.id);
        },
        onSuccess: () => {
            toast.success('模块重启命令已发送');
            queryClient.invalidateQueries({ queryKey: ['devices'] });
        },
        onError: (error) => {
            console.error('操作失败:', error);
            toast.error('操作失败');
        },
    });

    const handleSendSMS = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!to || !content) {
            toast.warning('请输入手机号和短信内容');
            return;
        }
        sendSMSMutation.mutate({ to, content });
    };

    if (!device) return null;

    // 获取设备状态配置
    const deviceStatus = device.status === 'online' ? STATUS_CONFIG.online : STATUS_CONFIG.offline;
    const flymodeStatus = FLYMODE_CONFIG[device.flymode.toString()];

    return (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 左侧：移动网络信息 */}
            <Card className="flex flex-col h-full">
                <CardHeader className="pb-4">
                    <CardTitle className="flex items-center gap-2 text-lg">
                        <Signal className="w-5 h-5 text-blue-600" />
                        移动网络信息
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 overflow-y-auto space-y-4">
                    {device.status === 'online' ? (
                        <div className="space-y-4">
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">SIM 状态</span>
                                <span className={`text-sm font-medium flex items-center gap-1 ${STATUS_CONFIG.online.color}`}>
                                    <div className={`w-2 h-2 rounded-full ${STATUS_CONFIG.online.dotColor}`} />
                                    正常
                                </span>
                            </div>
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">运营商</span>
                                <span className="text-sm font-medium">
                                    {device.operator || '未知'}
                                </span>
                            </div>
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">信号强度</span>
                                <div className="flex items-center gap-2">
                                    <span className="text-sm font-medium">{device.signalLevel}</span>
                                    <SignalStrength level={device.signalLevel} showText={true} />
                                </div>
                            </div>
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">IMSI</span>
                                <span className="text-sm font-mono font-medium">
                                    {device.imsi || '未知'}
                                </span>
                            </div>
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">ICCID</span>
                                <span className="text-sm font-mono font-medium">
                                    {device.iccid || '未知'}
                                </span>
                            </div>

                            <div className="pt-2">
                                <div className="flex justify-between items-center mb-2">
                                    <span className="text-sm font-medium text-gray-500">手机号</span>
                                    {!isEditingPhone ? (
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-7 px-3 text-blue-600 hover:text-blue-700 hover:bg-blue-50 transition-all"
                                            onClick={() => {
                                                setTempPhone(device.phoneNumber || '');
                                                setIsEditingPhone(true);
                                            }}
                                        >
                                            <Edit2 className="w-3.5 h-3.5 mr-1.5" />
                                            编辑
                                        </Button>
                                    ) : (
                                        <div className="flex items-center gap-1">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-7 w-7 p-0 text-green-600 hover:text-green-700 hover:bg-green-50 transition-all"
                                                onClick={() => updateDeviceMutation.mutate({
                                                    id: device.id,
                                                    phoneNumber: tempPhone
                                                })}
                                            >
                                                <Check className="w-4 h-4" />
                                            </Button>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-7 w-7 p-0 text-red-600 hover:text-red-700 hover:bg-red-50 transition-all"
                                                onClick={() => setIsEditingPhone(false)}
                                            >
                                                <X className="w-4 h-4" />
                                            </Button>
                                        </div>
                                    )}
                                </div>

                                {isEditingPhone ? (
                                    <Input
                                        value={tempPhone}
                                        onChange={(e) => setTempPhone(e.target.value)}
                                        className="h-9 font-mono text-sm"
                                        placeholder="请输入手机号"
                                    />
                                ) : (
                                    <div className="font-mono text-sm bg-gray-50 p-3 rounded-lg break-all min-h-[2.5rem] flex items-center">
                                        {device.phoneNumber || <span className="text-gray-400 italic">未设置</span>}
                                    </div>
                                )}
                            </div>
                        </div>
                    ) : (
                        <div className="flex flex-col items-center justify-center h-full text-gray-400 py-12">
                            <Wifi className="w-16 h-16 mb-3 opacity-30" />
                            <p className="text-base font-medium">设备离线</p>
                            <p className="text-sm text-gray-500 mt-1">无法获取设备信息</p>
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* 右侧：发送短信 */}
            <Card className="flex flex-col h-full">
                <CardHeader className="pb-4">
                    <CardTitle className="flex items-center gap-2 text-lg">
                        <Send className="w-5 h-5 text-green-600" />
                        发送短信
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 flex flex-col">
                    <form onSubmit={handleSendSMS} className="flex flex-col h-full space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                目标手机号
                            </label>
                            <Input
                                type="tel"
                                value={to}
                                onChange={(e) => setTo(e.target.value)}
                                placeholder="请输入手机号"
                                className="h-10"
                                required
                            />
                        </div>
                        <div className="flex-1 flex flex-col min-h-0">
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                短信内容
                            </label>
                            <Textarea
                                value={content}
                                onChange={(e) => setContent(e.target.value)}
                                placeholder="请输入短信内容"
                                className="flex-1 resize-none min-h-[100px] p-3"
                                required
                            />
                        </div>
                        <div className="text-sm text-gray-500">
                            通过 {device.name || device.serialPort} 发送
                        </div>
                        <Button
                            type="submit"
                            disabled={sendSMSMutation.isPending || device.status !== 'online'}
                            className={`w-full h-10 font-medium transition-all ${
                                sendSMSMutation.isPending || device.status !== 'online'
                                    ? 'bg-gray-400 cursor-not-allowed'
                                    : 'bg-green-600 hover:bg-green-700 text-white shadow-sm'
                            }`}
                        >
                            <Send className="w-4 h-4 mr-2" />
                            {sendSMSMutation.isPending ? '发送中...' : '发送短信'}
                        </Button>
                    </form>
                </CardContent>
            </Card>

            {/* 底部：设备状态 + 控制 */}
            <div className="lg:col-span-2 grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* 设备状态 */}
                <Card className="flex-1 flex flex-col h-full">
                    <CardHeader className="pb-4">
                        <CardTitle className="flex items-center gap-2 text-lg">
                            <Activity className="w-5 h-5 text-purple-600" />
                            设备状态
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1 overflow-y-auto space-y-4">
                        <div className="space-y-3">
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">连接状态</span>
                                <span className={`text-sm font-medium flex items-center gap-1 ${deviceStatus.color}`}>
                                    <div className={`w-2 h-2 rounded-full ${deviceStatus.dotColor}`} />
                                    {deviceStatus.text}
                                </span>
                            </div>
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">串口名称</span>
                                <span className="text-sm font-mono font-medium">{device.serialPort}</span>
                            </div>
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">设备名称</span>
                                <span className="text-sm font-medium">{device.name || '-'}</span>
                            </div>
                            {device.groupName && (
                                <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                    <span className="text-sm font-medium text-gray-500">设备分组</span>
                                    <span className="text-sm font-medium">{device.groupName}</span>
                                </div>
                            )}
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">最后心跳</span>
                                <span className="text-sm font-medium">
                                    {device.lastSeenAt > 0
                                        ? new Date(device.lastSeenAt).toLocaleString('zh-CN')
                                        : '-'}
                                </span>
                            </div>
                            <div className="flex justify-between items-center pb-3 border-b border-gray-200">
                                <span className="text-sm font-medium text-gray-500">飞行模式</span>
                                <span className={`text-sm font-medium ${flymodeStatus.color}`}>
                                    {flymodeStatus.text}
                                </span>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* 设备控制 */}
                <Card className="flex-1 flex flex-col h-full">
                    <CardHeader className="pb-4">
                        <CardTitle className="flex items-center gap-2 text-lg">
                            <RotateCcw className="w-5 h-5 text-orange-600" />
                            设备控制
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <div className="flex items-center justify-between p-3 rounded-lg bg-gray-50">
                                <span className="text-sm font-medium text-gray-700">飞行模式状态</span>
                                <span className={`px-3 py-1 rounded-full text-sm font-medium ${flymodeStatus.bgColor} ${flymodeStatus.color}`}>
                                    {flymodeStatus.text}
                                </span>
                            </div>

                            <div className="grid grid-cols-2 gap-3">
                                <Button
                                    onClick={() => setFlymodeMutation.mutate(true)}
                                    disabled={setFlymodeMutation.isPending || device.status !== 'online'}
                                    variant="outline"
                                    className={`h-10 font-medium transition-all ${
                                        setFlymodeMutation.isPending || device.status !== 'online'
                                            ? 'border-gray-300 text-gray-400 cursor-not-allowed'
                                            : 'border-orange-300 text-orange-700 hover:bg-orange-50'
                                    }`}
                                >
                                    开启飞行模式
                                </Button>
                                <Button
                                    onClick={() => setFlymodeMutation.mutate(false)}
                                    disabled={setFlymodeMutation.isPending || device.status !== 'online'}
                                    variant="outline"
                                    className={`h-10 font-medium transition-all ${
                                        setFlymodeMutation.isPending || device.status !== 'online'
                                            ? 'border-gray-300 text-gray-400 cursor-not-allowed'
                                            : 'border-green-300 text-green-700 hover:bg-green-50'
                                    }`}
                                >
                                    关闭飞行模式
                                </Button>
                            </div>

                            <div className="border-t border-gray-200 pt-4">
                                <Button
                                    onClick={() => rebootMcuMutation.mutate()}
                                    disabled={rebootMcuMutation.isPending || device.status !== 'online'}
                                    variant="outline"
                                    className={`w-full h-10 font-medium transition-all ${
                                        rebootMcuMutation.isPending || device.status !== 'online'
                                            ? 'border-gray-300 text-gray-400 cursor-not-allowed'
                                            : 'border-orange-300 text-orange-700 hover:bg-orange-50'
                                    }`}
                                >
                                    <RotateCcw className="w-4 h-4 mr-2" />
                                    重启模块
                                </Button>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
