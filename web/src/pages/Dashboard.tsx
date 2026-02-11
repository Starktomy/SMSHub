import { useQuery } from "@tanstack/react-query";
import { Globe, MessageSquare, Router, Signal, TrendingUp, Wifi, WifiOff } from 'lucide-react';
import { getStats } from '../api/messages';
import { devicesApi } from "@/api/devices";
import type { Device } from "@/api/devices";
import { StatCard } from "@/components/StatsCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { SignalStrength } from "@/components/SignalStrength";

export default function Dashboard() {
    // 获取统计信息 - 每 30 秒自动刷新
    const { data: stats, isLoading: statsLoading } = useQuery({
        queryKey: ['stats'],
        queryFn: getStats,
        refetchInterval: 30000,
    });

    // 获取设备列表 - 每 10 秒自动刷新
    const { data: devices = [], isLoading: devicesLoading } = useQuery<Device[]>({
        queryKey: ['devices'],
        queryFn: devicesApi.list,
        refetchInterval: 10000,
    });

    const loading = statsLoading || devicesLoading;

    // 统计设备数量
    const onlineDevices = devices.filter(d => d.status === 'online');

    // 获取信号最好的设备
    const bestSignalDevice = onlineDevices.length > 0
        ? onlineDevices.reduce((best, current) =>
            current.signalLevel > best.signalLevel ? current : best
        )
        : null;

    if (loading && !stats && devices.length === 0) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <h1 className="text-2xl font-bold text-gray-900">统计面板</h1>

            {/* 概览统计 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    label="在线设备"
                    value={onlineDevices.length}
                    icon={Router}
                    colorClass="bg-green-100 text-green-600"
                    subValue={`共 ${devices.length} 个设备`}
                />
                <StatCard
                    label="信号最强"
                    value={bestSignalDevice ? Math.round((bestSignalDevice.signalLevel / 31) * 100) : 0}
                    unit="%"
                    icon={Signal}
                    colorClass="bg-blue-100 text-blue-600"
                    subValue={bestSignalDevice ? (
                        <div className="flex items-center gap-1">
                            <span>{bestSignalDevice.name || bestSignalDevice.serialPort}</span>
                            <span>•</span>
                            <SignalStrength level={bestSignalDevice.signalLevel} className="scale-75 origin-left" />
                        </div>
                    ) : '无在线设备'}
                />
                <StatCard
                    label="总短信数"
                    value={stats?.totalCount || 0}
                    icon={MessageSquare}
                    colorClass="bg-purple-100 text-purple-600"
                    subValue={undefined}
                />
                <StatCard
                    label="今日短信"
                    value={stats?.todayCount || 0}
                    icon={TrendingUp}
                    colorClass="bg-orange-100 text-orange-600"
                    subValue={undefined}
                />
            </div>

            {/* 设备列表 */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Router className="w-5 h-5" />
                        设备状态
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    {devices.length === 0 ? (
                        <div className="text-center text-gray-400 py-8">
                            暂无设备，请前往「设备管理」添加设备
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {devices.map((device) => (
                                <div
                                    key={device.id}
                                    className={cn(
                                        "p-4 rounded-lg border transition-all",
                                        device.status === 'online'
                                            ? 'bg-green-50 border-green-200'
                                            : 'bg-gray-50 border-gray-200'
                                    )}
                                >
                                    <div className="flex items-center justify-between mb-2">
                                        <div className="flex items-center gap-2">
                                            {device.status === 'online' ? (
                                                <Wifi className="w-4 h-4 text-green-500" />
                                            ) : (
                                                <WifiOff className="w-4 h-4 text-gray-400" />
                                            )}
                                            <span className="font-medium truncate max-w-[120px]" title={device.name || device.serialPort}>
                                                {device.name || device.serialPort}
                                            </span>
                                        </div>
                                        <span className={cn(
                                            "text-xs px-2 py-1 rounded-full",
                                            device.status === 'online'
                                                ? 'bg-green-100 text-green-700'
                                                : 'bg-gray-100 text-gray-500'
                                        )}>
                                            {device.status === 'online' ? '在线' : '离线'}
                                        </span>
                                    </div>
                                    <div className="text-sm text-gray-500 space-y-1">
                                        {device.operator && (
                                            <div className="flex items-center gap-1">
                                                <Globe className="w-3 h-3" />
                                                <span>{device.operator}</span>
                                            </div>
                                        )}
                                        {device.phoneNumber && (
                                            <div className="text-xs text-gray-400">
                                                {device.phoneNumber}
                                            </div>
                                        )}
                                        {device.status === 'online' && (
                                            <div className="flex items-center gap-1 mt-1">
                                                <SignalStrength level={device.signalLevel} showText={true} />
                                            </div>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* 系统信息 */}
            <Card>
                <CardHeader>
                    <CardTitle>系统信息</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="space-y-2 text-sm text-gray-600">
                        <p>• 支持多设备集中管理，实时监控设备状态</p>
                        <p>• 自动接收短信并发送通知到配置的渠道</p>
                        <p>• 自动接收来电并发送通知</p>
                        <p>• 支持定时发送短信，可指定执行设备</p>
                        <p>• 支持批量发送，自动负载均衡</p>
                        <p>• 短信记录自动保存到数据库</p>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
