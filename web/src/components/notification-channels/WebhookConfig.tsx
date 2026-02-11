import { Link, CheckCircle2, TestTube } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface WebhookConfigProps {
  enabled: boolean;
  url: string;
  method: string;
  contentType: string;
  headers: string;
  body: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onUpdate: (field: string, value: any) => void;
  onTest: () => void;
  isTestPending: boolean;
}

export function WebhookConfig({
  enabled,
  url,
  method,
  contentType,
  headers,
  body,
  onUpdate,
  onTest,
  isTestPending,
}: WebhookConfigProps) {
  return (
    <Card
      className={`border transition-all ${
        enabled
          ? 'border-orange-200 bg-gradient-to-br from-white to-orange-50/20'
          : 'border-gray-200 opacity-95'
      }`}
    >
      <CardHeader className="border-b border-gray-100 bg-white/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3 flex-1">
            <div
              className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                enabled ? 'bg-orange-50 text-orange-600' : 'bg-gray-100 text-gray-400'
              }`}
            >
              <Link size={24} />
            </div>
            <div className="flex-1">
              <div className="flex items-center space-x-2">
                <CardTitle className="text-lg font-bold text-gray-800">自定义 Webhook</CardTitle>
                <div
                  className={`w-2 h-2 rounded-full ${enabled ? 'bg-green-500' : 'bg-gray-300'}`}
                ></div>
                <span className="text-xs text-gray-500">
                  {enabled ? '已启用' : '未启用'}
                </span>
              </div>
              <CardDescription className="mt-1.5 text-xs">
                配置自定义 HTTP 回调接口接收短信通知
              </CardDescription>
            </div>
          </div>
          <div className="flex items-center space-x-3">
            {enabled && (
              <Button
                variant="outline"
                size="sm"
                disabled={isTestPending}
                onClick={onTest}
                className="text-xs bg-gray-100 hover:bg-gray-200 transition-colors border-none cursor-pointer"
              >
                <TestTube className="w-3.5 h-3.5 mr-1.5" />
                {isTestPending ? '测试中...' : '发送测试'}
              </Button>
            )}
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                className="sr-only peer"
                checked={enabled}
                onChange={(e) => onUpdate('webhookEnabled', e.target.checked)}
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-orange-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-orange-600"></div>
            </label>
          </div>
        </div>
      </CardHeader>

      {enabled && (
        <CardContent className="space-y-4 animate-in slide-in-from-top-2 duration-200 pt-6">
          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              Webhook URL <span className="text-red-500">*</span>
            </label>
            <Input
              value={url}
              onChange={(e) => onUpdate('webhookUrl', e.target.value)}
              placeholder="https://your-server.com/webhook"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-orange-500 focus:ring-1 focus:ring-orange-500 transition-all font-mono text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              HTTP 方法
            </label>
            <Select
              value={method}
              onValueChange={(value) => onUpdate('webhookMethod', value)}
            >
              <SelectTrigger className="bg-gray-50 border-gray-200 focus:bg-white focus:border-orange-500 focus:ring-1 focus:ring-orange-500 transition-all">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GET">GET</SelectItem>
                <SelectItem value="POST">POST</SelectItem>
                <SelectItem value="PUT">PUT</SelectItem>
                <SelectItem value="PATCH">PATCH</SelectItem>
                <SelectItem value="DELETE">DELETE</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              请求类型 <span className="text-red-500">*</span>
            </label>
            <Input
              value={contentType || 'application/json; charset=utf-8'}
              onChange={(e) => onUpdate('webhookContentType', e.target.value)}
              placeholder="application/json; charset=utf-8"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              请求体模板 <span className="text-red-500">*</span>
            </label>
            <Textarea
              value={body}
              onChange={(e) => onUpdate('webhookBody', e.target.value)}
              placeholder='{"from": "{{from}}", "content": "{{content}}", "timestamp": "{{timestamp}}"}'
              rows={6}
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-orange-500 focus:ring-1 focus:ring-orange-500 transition-all font-mono text-xs"
            />
            <p className="text-xs text-gray-400 mt-1.5">
              支持模板变量：<code className="bg-gray-200 px-1 py-0.5 rounded">{'{{from}}'}</code>（发送方）、
              <code className="bg-gray-200 px-1 py-0.5 rounded">{'{{content}}'}</code>（短信内容）、
              <code className="bg-gray-200 px-1 py-0.5 rounded">{'{{timestamp}}'}</code>（时间戳）
            </p>
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              自定义请求头 (JSON 格式)
            </label>
            <Textarea
              value={headers}
              onChange={(e) => onUpdate('webhookHeaders', e.target.value)}
              placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'
              rows={4}
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-orange-500 focus:ring-1 focus:ring-orange-500 transition-all font-mono text-xs"
            />
            <p className="text-xs text-gray-400 mt-1.5">
              可选，格式为 JSON 对象，例如: {`{"key": "value"}`}
            </p>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="text-xs font-bold text-blue-900 mb-2 flex items-center gap-1.5">
              <CheckCircle2 size={14} />
              模板变量说明
            </div>
            <div className="text-xs text-blue-800 space-y-2">
              <p>请求体支持以下模板变量：</p>
              <ul className="list-disc list-inside space-y-1 ml-2">
                <li><code className="bg-white px-1.5 py-0.5 rounded border border-blue-200">{'{{from}}'}</code> - 短信发送方手机号</li>
                <li><code className="bg-white px-1.5 py-0.5 rounded border border-blue-200">{'{{content}}'}</code> - 短信内容</li>
                <li><code className="bg-white px-1.5 py-0.5 rounded border border-blue-200">{'{{timestamp}}'}</code> - 接收时间（格式：2006-01-02 15:04:05）</li>
              </ul>
              <p className="mt-2">示例模板：</p>
              <pre className="bg-white border border-blue-100 rounded p-3 mt-2 overflow-x-auto text-[11px] leading-relaxed">
{`{
  "from": "{{from}}",
  "content": "{{content}}",
  "timestamp": "{{timestamp}}"
}`}
              </pre>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}
