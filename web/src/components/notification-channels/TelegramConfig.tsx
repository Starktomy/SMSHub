import { Bell, Shield, TestTube } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface TelegramConfigProps {
  enabled: boolean;
  apiToken: string;
  userid: string;
  proxyEnabled: boolean;
  proxyUrl: string;
  proxyUsername?: string;
  proxyPassword?: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onUpdate: (field: string, value: any) => void;
  onTest: () => void;
  isTestPending: boolean;
}

export function TelegramConfig({
  enabled,
  apiToken,
  userid,
  proxyEnabled,
  proxyUrl,
  proxyUsername,
  proxyPassword,
  onUpdate,
  onTest,
  isTestPending,
}: TelegramConfigProps) {
  return (
    <Card
      className={`border transition-all ${
        enabled
          ? 'border-blue-200 bg-gradient-to-br from-white to-blue-50/20'
          : 'border-gray-200 opacity-95'
      }`}
    >
      <CardHeader className="border-b border-gray-100 bg-white/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3 flex-1">
            <div
              className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                enabled ? 'bg-blue-50 text-blue-600' : 'bg-gray-100 text-gray-400'
              }`}
            >
              <Bell size={24} />
            </div>
            <div className="flex-1">
              <div className="flex items-center space-x-2">
                <CardTitle className="text-lg font-bold text-gray-800">Telegram通知</CardTitle>
                <div
                  className={`w-2 h-2 rounded-full ${enabled ? 'bg-green-500' : 'bg-gray-300'}`}
                ></div>
                <span className="text-xs text-gray-500">
                  {enabled ? '已启用' : '未启用'}
                </span>
                <div
                  className={`w-2 h-2 rounded-full ${proxyEnabled ? 'bg-green-500' : 'bg-gray-300'}`}
                ></div>
                <span className="text-xs text-gray-500">
                  {proxyEnabled ? '代理开' : '代理关'}
                </span>
              </div>
              <CardDescription className="mt-1.5 text-xs">
                了解更多：
                <a
                  href="https://core.telegram.org/bots/api"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700 hover:underline ml-1 transition-colors font-medium"
                >
                  Telegram自定义机器人接入文档
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
            {enabled && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onUpdate('telegramProxyEnabled', !proxyEnabled)}
                className={`text-xs border-none transition-colors ${
                  proxyEnabled
                    ? 'bg-blue-100 text-blue-700'
                    : 'bg-gray-100 text-gray-600'
                }`}
              >
                🌐 HTTP代理
              </Button>
            )}
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                className="sr-only peer"
                checked={enabled}
                onChange={(e) => onUpdate('telegramlEnabled', e.target.checked)}
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>
        </div>
      </CardHeader>

      {enabled && (
        <CardContent className="space-y-4 animate-in slide-in-from-top-2 duration-200 pt-6">
          {proxyEnabled && (
            <div className="space-y-3 rounded-lg border border-blue-100 bg-blue-50/40 p-3 animate-in fade-in duration-200">
              <div>
                <label className="block text-xs font-semibold text-gray-600 mb-1 uppercase tracking-wide">
                  HTTP代理地址
                </label>
                <Input
                  value={proxyUrl}
                  onChange={(e) => onUpdate('telegramProxyUrl', e.target.value)}
                  placeholder="http://127.0.0.1:7890"
                  className="font-mono text-sm"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">
                    用户名（可选）
                  </label>
                  <Input
                    value={proxyUsername}
                    onChange={(e) => onUpdate('telegramProxyUsername', e.target.value)}
                    className="font-mono text-sm"
                  />
                </div>

                <div>
                  <label className="block text-xs text-gray-500 mb-1">
                    密码（可选）
                  </label>
                  <Input
                    type="password"
                    value={proxyPassword}
                    onChange={(e) => onUpdate('telegramProxyPassword', e.target.value)}
                    className="font-mono text-sm"
                  />
                </div>
              </div>
            </div>
          )}

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              API Token <span className="text-red-500">*</span>
            </label>
            <Input
              value={apiToken}
              onChange={(e) => onUpdate('telegramApiToken', e.target.value)}
              placeholder="apiToken"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all font-mono text-sm"
            />
          </div>
          <p className="text-xs text-gray-400 mt-1.5">使用@BotFather机器人获取</p>
          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              用户 ID
            </label>
            <div className="relative">
              <Input
                value={userid}
                onChange={(e) => onUpdate('telegramUserid', e.target.value)}
                placeholder="User ID"
                className="bg-gray-50 border-gray-200 focus:bg-white focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all font-mono text-sm pr-10"
              />
              <Shield
                size={14}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
            </div>
            <p className="text-xs text-gray-400 mt-1.5">使用@userinfobot机器人获取</p>
          </div>
        </CardContent>
      )}
    </Card>
  );
}
