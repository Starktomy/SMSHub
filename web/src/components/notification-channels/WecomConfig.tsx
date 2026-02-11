import { MessageSquare, TestTube } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface WecomConfigProps {
  enabled: boolean;
  secretKey: string;
  onUpdate: (field: string, value: any) => void;
  onTest: () => void;
  isTestPending: boolean;
}

export function WecomConfig({
  enabled,
  secretKey,
  onUpdate,
  onTest,
  isTestPending,
}: WecomConfigProps) {
  return (
    <Card
      className={`border transition-all ${
        enabled
          ? 'border-green-200 bg-gradient-to-br from-white to-green-50/20'
          : 'border-gray-200 opacity-95'
      }`}
    >
      <CardHeader className="border-b border-gray-100 bg-white/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3 flex-1">
            <div
              className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                enabled ? 'bg-green-50 text-green-600' : 'bg-gray-100 text-gray-400'
              }`}
            >
              <MessageSquare size={24} />
            </div>
            <div className="flex-1">
              <div className="flex items-center space-x-2">
                <CardTitle className="text-lg font-bold text-gray-800">企业微信通知</CardTitle>
                <div
                  className={`w-2 h-2 rounded-full ${enabled ? 'bg-green-500' : 'bg-gray-300'}`}
                ></div>
                <span className="text-xs text-gray-500">
                  {enabled ? '已启用' : '未启用'}
                </span>
              </div>
              <CardDescription className="mt-1.5 text-xs">
                了解更多：
                <a
                  href="https://work.weixin.qq.com/api/doc/90000/90136/91770"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700 hover:underline ml-1 transition-colors font-medium"
                >
                  企业微信群机器人配置说明
                </a>
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
                onChange={(e) => onUpdate('wecomEnabled', e.target.checked)}
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-green-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-600"></div>
            </label>
          </div>
        </div>
      </CardHeader>

      {enabled && (
        <CardContent className="space-y-4 animate-in slide-in-from-top-2 duration-200 pt-6">
          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              Webhook Key <span className="text-red-500">*</span>
            </label>
            <Input
              value={secretKey}
              onChange={(e) => onUpdate('wecomSecretKey', e.target.value)}
              placeholder="企业微信群机器人的 Webhook Key"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-green-500 focus:ring-1 focus:ring-green-500 transition-all font-mono text-sm"
            />
          </div>
        </CardContent>
      )}
    </Card>
  );
}
