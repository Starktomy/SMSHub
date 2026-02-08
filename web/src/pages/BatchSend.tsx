import {useState} from 'react';
import {useQuery, useMutation} from '@tanstack/react-query';
import {devicesApi} from '@/api/devices';
import type {Device, BatchSendRequest, BatchSendResult} from '@/api/devices';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Textarea} from '@/components/ui/textarea';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {toast} from 'sonner';
import {Send, CheckCircle, XCircle, Loader2} from 'lucide-react';

export default function BatchSend() {
    const [recipients, setRecipients] = useState('');
    const [content, setContent] = useState('');
    const [deviceId, setDeviceId] = useState('auto');
    const [strategy, setStrategy] = useState<'auto' | 'round_robin' | 'random' | 'signal_best'>('auto');
    const [results, setResults] = useState<BatchSendResult[]>([]);

    // 获取设备列表
    const {data: devices} = useQuery({
        queryKey: ['devices'],
        queryFn: devicesApi.list,
    });

    // 批量发送
    const batchSendMutation = useMutation({
        mutationFn: (data: BatchSendRequest) => devicesApi.batchSend(data),
        onSuccess: (data) => {
            setResults(data.results || []);
            const successCount = data.results?.filter((r) => r.success).length || 0;
            const failCount = (data.results?.length || 0) - successCount;
            if (failCount === 0) {
                toast.success(`全部发送成功 (${successCount} 条)`);
            } else {
                toast.warning(`发送完成: ${successCount} 成功, ${failCount} 失败`);
            }
        },
        onError: () => {
            toast.error('批量发送失败');
        },
    });

    const handleSend = () => {
        const recipientList = recipients
            .split('\n')
            .map((r) => r.trim())
            .filter((r) => r.length > 0);

        if (recipientList.length === 0) {
            toast.error('请输入收件人号码');
            return;
        }

        if (!content.trim()) {
            toast.error('请输入短信内容');
            return;
        }

        const request: BatchSendRequest = {
            recipients: recipientList,
            content: content.trim(),
            strategy,
        };

        if (deviceId && deviceId !== 'auto') {
            request.deviceId = deviceId;
        }

        batchSendMutation.mutate(request);
    };

    const onlineDevices = devices?.filter((d) => d.status === 'online') || [];

    return (
        <div className="space-y-6">
            {/* 头部 */}
            <div>
                <h1 className="text-2xl font-bold text-gray-900">批量发送</h1>
                <p className="text-sm text-gray-500 mt-1">
                    向多个号码批量发送短信，支持多设备负载均衡
                </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* 发送表单 */}
                <Card>
                    <CardHeader>
                        <CardTitle>发送设置</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {/* 收件人 */}
                        <div>
                            <label className="text-sm font-medium">
                                收件人号码 <span className="text-gray-400">(每行一个)</span>
                            </label>
                            <Textarea
                                placeholder="+8613800138000&#10;+8613900139000&#10;..."
                                rows={6}
                                value={recipients}
                                onChange={(e) => setRecipients(e.target.value)}
                            />
                            <p className="text-xs text-gray-400 mt-1">
                                已输入 {recipients.split('\n').filter((r) => r.trim()).length} 个号码
                            </p>
                        </div>

                        {/* 短信内容 */}
                        <div>
                            <label className="text-sm font-medium">短信内容</label>
                            <Textarea
                                placeholder="输入短信内容..."
                                rows={4}
                                value={content}
                                onChange={(e) => setContent(e.target.value)}
                            />
                            <p className="text-xs text-gray-400 mt-1">
                                {content.length} 字符
                            </p>
                        </div>

                        {/* 设备选择 */}
                        <div>
                            <label className="text-sm font-medium">发送设备</label>
                            <Select value={deviceId} onValueChange={setDeviceId}>
                                <SelectTrigger>
                                    <SelectValue placeholder="自动选择设备" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="auto">自动选择</SelectItem>
                                    {onlineDevices.map((device) => (
                                        <SelectItem key={device.id} value={device.id}>
                                            {device.name || device.serialPort} {device.operator ? `(${device.operator})` : ''}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        {/* 负载均衡策略 */}
                        {deviceId === 'auto' && (
                            <div>
                                <label className="text-sm font-medium">分配策略</label>
                                <Select
                                    value={strategy}
                                    onValueChange={(v) => setStrategy(v as typeof strategy)}
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="auto">自动 (优先信号最强)</SelectItem>
                                        <SelectItem value="round_robin">轮询</SelectItem>
                                        <SelectItem value="random">随机</SelectItem>
                                        <SelectItem value="signal_best">信号最强</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        )}

                        {/* 发送按钮 */}
                        <Button
                            className="w-full"
                            onClick={handleSend}
                            disabled={batchSendMutation.isPending}
                        >
                            {batchSendMutation.isPending ? (
                                <>
                                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                    发送中...
                                </>
                            ) : (
                                <>
                                    <Send className="w-4 h-4 mr-2" />
                                    批量发送
                                </>
                            )}
                        </Button>

                        {/* 在线设备数 */}
                        <p className="text-xs text-center text-gray-400">
                            当前在线设备: {onlineDevices.length} 个
                        </p>
                    </CardContent>
                </Card>

                {/* 发送结果 */}
                <Card>
                    <CardHeader>
                        <CardTitle>发送结果</CardTitle>
                    </CardHeader>
                    <CardContent>
                        {results.length === 0 ? (
                            <div className="text-center text-gray-400 py-12">
                                发送后将在此显示结果
                            </div>
                        ) : (
                            <div className="space-y-2 max-h-96 overflow-y-auto">
                                {results.map((result, index) => (
                                    <div
                                        key={index}
                                        className={`flex items-center justify-between p-3 rounded-lg ${
                                            result.success
                                                ? 'bg-green-50 border border-green-200'
                                                : 'bg-red-50 border border-red-200'
                                        }`}
                                    >
                                        <div className="flex items-center gap-2">
                                            {result.success ? (
                                                <CheckCircle className="w-4 h-4 text-green-500" />
                                            ) : (
                                                <XCircle className="w-4 h-4 text-red-500" />
                                            )}
                                            <span className="font-mono text-sm">
                                                {result.recipient}
                                            </span>
                                        </div>
                                        <div className="text-xs text-gray-500">
                                            {result.success ? (
                                                <span className="text-green-600">已发送</span>
                                            ) : (
                                                <span className="text-red-600">{result.error}</span>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}

                        {/* 统计 */}
                        {results.length > 0 && (
                            <div className="mt-4 pt-4 border-t flex justify-center gap-6 text-sm">
                                <div className="text-green-600">
                                    成功: {results.filter((r) => r.success).length}
                                </div>
                                <div className="text-red-600">
                                    失败: {results.filter((r) => !r.success).length}
                                </div>
                                <div className="text-gray-500">
                                    总计: {results.length}
                                </div>
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
