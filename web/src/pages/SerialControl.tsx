import {useState} from 'react';
import {Activity, RotateCcw, Send, Signal, Wifi, Router} from 'lucide-react';
import {toast} from 'sonner';
import {useMutation, useQuery} from '@tanstack/react-query';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Button} from '@/components/ui/button';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {devicesApi} from '@/api/devices';
import type {Device} from '@/api/devices';
import { SignalStrength } from '@/components/SignalStrength';

export default function SerialControl() {
    const [to, setTo] = useState('');
    const [content, setContent] = useState('');
    const [selectedDeviceId, setSelectedDeviceId] = useState<string>('');

    // 获取设备列表 - 每 10 秒自动刷新
    const {data: devices = [], refetch: refetchDevices} = useQuery<Device[]>({
        queryKey: ['devices'],
        queryFn: devicesApi.list,
        refetchInterval: 10000,
    });

    // 在线设备列表
    const onlineDevices = devices.filter(d => d.status === 'online');

    // 当前选中的设备
    const selectedDevice = devices.find(d => d.id === selectedDeviceId);

    // 自动选择第一个在线设备
    if (!selectedDeviceId && onlineDevices.length > 0) {
        setSelectedDeviceId(onlineDevices[0].id);
    }

    // 发送短信 Mutation
    const sendSMSMutation = useMutation({
        mutationFn: (data: { to: string; content: string }) => {
            if (selectedDeviceId) {
                return devicesApi.sendSMS(selectedDeviceId, data.to, data.content);
            }
            return devicesApi.autoSendSMS(data.to, data.content);
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
            if (!selectedDeviceId) {
                return Promise.reject(new Error('请先选择设备'));
            }
            return devicesApi.setFlymode(selectedDeviceId, enabled);
        },
        onSuccess: () => {
            toast.success('设置成功');
            refetchDevices();
        },
        onError: (error) => {
            console.error('操作失败:', error);
            toast.error('操作失败');
        },
    });

    // 重启模块 Mutation
    const rebootMcuMutation = useMutation({
        mutationFn: () => {
            if (!selectedDeviceId) {
                return Promise.reject(new Error('请先选择设备'));
            }
            return devicesApi.reboot(selectedDeviceId);
        },
        onSuccess: () => {
            toast.success('模块重启命令已发送');
            refetchDevices();
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
        sendSMSMutation.mutate({to, content});
    };

    return (
        <div className="flex flex-col overflow-hidden">
            {/* 顶部标题和设备选择 */}
            <div className="mb-6 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                <h1 className="text-2xl font-bold text-gray-900">设备控制</h1>
                <div className="flex items-center gap-3">
                    <Router className="w-4 h-4 text-gray-400" />
                    <Select value={selectedDeviceId} onValueChange={setSelectedDeviceId}>
                        <SelectTrigger className="w-64">
                            <SelectValue placeholder="选择设备" />
                        </SelectTrigger>
                        <SelectContent>
                            {devices.length === 0 ? (
                                <SelectItem value="none" disabled>暂无设备</SelectItem>
                            ) : (
                                devices.map((device) => (
                                    <SelectItem key={device.id} value={device.id}>
                                        <div className="flex items-center gap-2">
                                            <span className={`w-2 h-2 rounded-full ${
                                                device.status === 'online' ? 'bg-green-500' : 'bg-gray-300'
                                            }`} />
                                            {device.name || device.serialPort}
                                            {device.operator && ` (${device.operator})`}
                                        </div>
                                    </SelectItem>
                                ))
                            )}
                        </SelectContent>
                    </Select>
                    <span className="text-xs text-gray-400">
                        在线 {onlineDevices.length} / {devices.length}
                    </span>
                </div>
            </div>

            {/* 未选择设备时的提示 */}
            {!selectedDevice && (
                <div className="flex flex-col items-center justify-center py-20 bg-white rounded-xl border border-gray-200">
                    <Router className="w-16 h-16 text-gray-300 mb-4" />
                    <p className="text-gray-500 font-medium">请选择一个设备</p>
                    <p className="text-gray-400 text-sm mt-1">从上方下拉菜单中选择要控制的设备</p>
                </div>
            )}

            {/* 主内容区 - 三列布局 */}
            {selectedDevice && (
                <div className="flex-1 grid grid-cols-1 lg:grid-cols-3 gap-4 min-h-0">
                    {/* 左侧：移动网络信息 */}
                    <Card className="flex flex-col min-h-0">
                        <CardHeader className="pb-3">
                            <CardTitle className="flex items-center gap-2 text-base">
                                <Signal className="w-4 h-4 text-blue-600"/>
                                移动网络信息
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 overflow-y-auto">
                            {selectedDevice.status === 'online' ? (
                                <div className="space-y-3">
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">SIM 状态</span>
                                        <span className="text-sm font-medium">
                                            <span className="text-green-600 flex items-center gap-1">
                                                <div className="w-1.5 h-1.5 rounded-full bg-green-600"></div>
                                                正常
                                            </span>
                                        </span>
                                    </div>
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">运营商</span>
                                        <span className="text-sm font-medium">
                                            {selectedDevice.operator || '未知'}
                                        </span>
                                    </div>
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">信号强度</span>
                                        <div className="flex items-center gap-2">
                                            <span className="text-sm font-medium">{selectedDevice.signalLevel}</span>
                                            <SignalStrength level={selectedDevice.signalLevel} showText={true} />
                                        </div>
                                    </div>
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">信号百分比</span>
                                        <span className="text-sm font-medium">
                                            {Math.round((selectedDevice.signalLevel / 31) * 100)}%
                                        </span>
                                    </div>
                                    {selectedDevice.phoneNumber && (
                                        <div className="pt-1">
                                            <div className="text-xs text-gray-500 mb-1">手机号</div>
                                            <div className="font-mono text-xs bg-gray-50 p-1.5 rounded break-all">
                                                {selectedDevice.phoneNumber}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ) : (
                                <div className="flex flex-col items-center justify-center h-full text-gray-400">
                                    <Wifi className="w-12 h-12 mb-2 opacity-30"/>
                                    <p className="text-sm">设备离线</p>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* 中间：发送短信 */}
                    <Card className="flex flex-col min-h-0">
                        <CardHeader className="pb-3">
                            <CardTitle className="flex items-center gap-2 text-base">
                                <Send className="w-4 h-4 text-green-600"/>
                                发送短信
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 flex flex-col">
                            <form onSubmit={handleSendSMS} className="flex flex-col h-full space-y-3">
                                <div>
                                    <label className="block text-xs font-medium text-gray-700 mb-1.5">
                                        目标手机号
                                    </label>
                                    <Input
                                        type="tel"
                                        value={to}
                                        onChange={(e) => setTo(e.target.value)}
                                        placeholder="请输入手机号"
                                        className="h-9"
                                        required
                                    />
                                </div>
                                <div className="flex-1 flex flex-col min-h-0">
                                    <label className="block text-xs font-medium text-gray-700 mb-1.5">
                                        短信内容
                                    </label>
                                    <Textarea
                                        value={content}
                                        onChange={(e) => setContent(e.target.value)}
                                        placeholder="请输入短信内容"
                                        className="flex-1 resize-none"
                                        required
                                    />
                                </div>
                                <div className="text-xs text-gray-400">
                                    通过 {selectedDevice.name || selectedDevice.serialPort} 发送
                                </div>
                                <Button
                                    type="submit"
                                    disabled={sendSMSMutation.isPending || selectedDevice.status !== 'online'}
                                    className="w-full bg-green-600 hover:bg-green-700 h-9"
                                >
                                    <Send className="w-3.5 h-3.5 mr-2"/>
                                    {sendSMSMutation.isPending ? '发送中...' : '发送短信'}
                                </Button>
                            </form>
                        </CardContent>
                    </Card>

                    {/* 右侧：设备状态 + 控制 */}
                    <div className="flex flex-col gap-4 min-h-0">
                        {/* 设备状态 */}
                        <Card className="flex-1 flex flex-col min-h-0 gap-2">
                            <CardHeader className="pb-3">
                                <CardTitle className="flex items-center gap-2 text-base">
                                    <Activity className="w-4 h-4 text-purple-600"/>
                                    设备状态
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="flex-1 overflow-y-auto">
                                <div className="space-y-2">
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">连接状态</span>
                                        <span className="text-sm font-medium">
                                            {selectedDevice.status === 'online' ? (
                                                <span className="text-green-600 flex items-center gap-1">
                                                    <div className="w-1.5 h-1.5 rounded-full bg-green-600"></div>
                                                    在线
                                                </span>
                                            ) : (
                                                <span className="text-red-600 flex items-center gap-1">
                                                    <div className="w-1.5 h-1.5 rounded-full bg-red-600"></div>
                                                    离线
                                                </span>
                                            )}
                                        </span>
                                    </div>
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">串口名称</span>
                                        <span className="text-sm font-medium font-mono">{selectedDevice.serialPort}</span>
                                    </div>
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">设备名称</span>
                                        <span className="text-sm font-medium">{selectedDevice.name || '-'}</span>
                                    </div>
                                    {selectedDevice.groupName && (
                                        <div className="flex justify-between items-center pb-2 border-b">
                                            <span className="text-xs text-gray-500">设备分组</span>
                                            <span className="text-sm font-medium">{selectedDevice.groupName}</span>
                                        </div>
                                    )}
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">最后心跳</span>
                                        <span className="text-sm font-medium">
                                            {selectedDevice.lastSeenAt > 0
                                                ? new Date(selectedDevice.lastSeenAt).toLocaleString('zh-CN')
                                                : '-'}
                                        </span>
                                    </div>
                                    <div className="flex justify-between items-center pb-2 border-b">
                                        <span className="text-xs text-gray-500">飞行模式</span>
                                        <span className="text-sm font-medium">
                                            {selectedDevice.flymode ? (
                                                <span className="text-orange-600">已启用</span>
                                            ) : (
                                                <span className="text-green-600">已禁用</span>
                                            )}
                                        </span>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>

                        {/* 设备控制 */}
                        <Card className={'gap-2'}>
                            <CardHeader className="pb-3">
                                <CardTitle className="flex items-center gap-2 text-base">
                                    <RotateCcw className="w-4 h-4 text-orange-600"/>
                                    设备控制
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-3">
                                    <p className="text-xs text-gray-600">
                                        飞行模式状态：{selectedDevice.flymode ? (
                                            <span className="text-orange-600 font-medium">已启用</span>
                                        ) : (
                                            <span className="text-green-600 font-medium">已禁用</span>
                                        )}
                                    </p>
                                    <div className="flex gap-2">
                                        <Button
                                            onClick={() => setFlymodeMutation.mutate(true)}
                                            disabled={setFlymodeMutation.isPending || selectedDevice.status !== 'online'}
                                            variant="outline"
                                            className="flex-1 border-orange-300 text-orange-700 hover:bg-orange-50 h-9 cursor-pointer"
                                        >
                                            开启飞行模式
                                        </Button>
                                        <Button
                                            onClick={() => setFlymodeMutation.mutate(false)}
                                            disabled={setFlymodeMutation.isPending || selectedDevice.status !== 'online'}
                                            variant="outline"
                                            className="flex-1 border-green-300 text-green-700 hover:bg-green-50 h-9 cursor-pointer"
                                        >
                                            关闭飞行模式
                                        </Button>
                                    </div>
                                    <div className="border-t pt-2">
                                        <Button
                                            onClick={() => rebootMcuMutation.mutate()}
                                            disabled={rebootMcuMutation.isPending || selectedDevice.status !== 'online'}
                                            variant="outline"
                                            className="w-full border-orange-300 text-orange-700 hover:bg-orange-50 h-9"
                                        >
                                            <RotateCcw className="w-3.5 h-3.5 mr-2"/>
                                            重启模块
                                        </Button>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            )}
        </div>
    );
}
