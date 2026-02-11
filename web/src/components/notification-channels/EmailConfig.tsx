import { Mail, Shield, CheckCircle2, TestTube } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface EmailConfigProps {
  enabled: boolean;
  smtpHost: string;
  smtpPort: string;
  username: string;
  password?: string;
  from: string;
  to: string;
  subject: string;
  onUpdate: (field: string, value: any) => void;
  onTest: () => void;
  isTestPending: boolean;
}

export function EmailConfig({
  enabled,
  smtpHost,
  smtpPort,
  username,
  password,
  from,
  to,
  subject,
  onUpdate,
  onTest,
  isTestPending,
}: EmailConfigProps) {
  return (
    <Card
      className={`border transition-all ${
        enabled
          ? 'border-indigo-200 bg-gradient-to-br from-white to-indigo-50/20'
          : 'border-gray-200 opacity-95'
      }`}
    >
      <CardHeader className="border-b border-gray-100 bg-white/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3 flex-1">
            <div
              className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                enabled ? 'bg-indigo-50 text-indigo-600' : 'bg-gray-100 text-gray-400'
              }`}
            >
              <Mail size={24} />
            </div>
            <div className="flex-1">
              <div className="flex items-center space-x-2">
                <CardTitle className="text-lg font-bold text-gray-800">邮件通知</CardTitle>
                <div
                  className={`w-2 h-2 rounded-full ${enabled ? 'bg-green-500' : 'bg-gray-300'}`}
                ></div>
                <span className="text-xs text-gray-500">
                  {enabled ? '已启用' : '未启用'}
                </span>
              </div>
              <CardDescription className="mt-1.5 text-xs">
                通过 SMTP 协议发送邮件通知
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
                onChange={(e) => onUpdate('emailEnabled', e.target.checked)}
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600"></div>
            </label>
          </div>
        </div>
      </CardHeader>

      {enabled && (
        <CardContent className="space-y-4 animate-in slide-in-from-top-2 duration-200 pt-6">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
                SMTP 服务器 <span className="text-red-500">*</span>
              </label>
              <Input
                value={smtpHost}
                onChange={(e) => onUpdate('emailSmtpHost', e.target.value)}
                placeholder="smtp.example.com"
                className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
                SMTP 端口 <span className="text-red-500">*</span>
              </label>
              <Input
                value={smtpPort}
                onChange={(e) => onUpdate('emailSmtpPort', e.target.value)}
                placeholder="587"
                className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              用户名 <span className="text-red-500">*</span>
            </label>
            <Input
              value={username}
              onChange={(e) => onUpdate('emailUsername', e.target.value)}
              placeholder="your-email@example.com"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              密码/授权码 <span className="text-red-500">*</span>
            </label>
            <div className="relative">
              <Input
                type="password"
                value={password}
                onChange={(e) => onUpdate('emailPassword', e.target.value)}
                placeholder="SMTP 密码或授权码"
                className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm pr-10"
              />
              <Shield
                size={14}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
            </div>
            <p className="text-xs text-gray-400 mt-1.5">部分邮箱服务商（如 QQ、163 等）需要使用授权码而非登录密码</p>
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              发件人地址 <span className="text-red-500">*</span>
            </label>
            <Input
              value={from}
              onChange={(e) => onUpdate('emailFrom', e.target.value)}
              placeholder="sender@example.com"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              收件人地址 <span className="text-red-500">*</span>
            </label>
            <Input
              value={to}
              onChange={(e) => onUpdate('emailTo', e.target.value)}
              placeholder="receiver@example.com（多个收件人用逗号分隔）"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-600 mb-2 uppercase tracking-wide">
              邮件主题模板 <span className="text-red-500">*</span>
            </label>
            <Input
              value={subject}
              onChange={(e) => onUpdate('emailSubject', e.target.value)}
              placeholder="收到新短信 - {{from}}"
              className="bg-gray-50 border-gray-200 focus:bg-white focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all font-mono text-sm"
            />
            <p className="text-xs text-gray-400 mt-1.5">
              支持模板变量：<code className="bg-gray-200 px-1 py-0.5 rounded">{'{{from}}'}</code>（发送方）、
              <code className="bg-gray-200 px-1 py-0.5 rounded">{'{{content}}'}</code>（短信内容）、
              <code className="bg-gray-200 px-1 py-0.5 rounded">{'{{timestamp}}'}</code>（时间戳）
            </p>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="text-xs font-bold text-blue-900 mb-2 flex items-center gap-1.5">
              <CheckCircle2 size={14} />
              常用邮箱 SMTP 配置
            </div>
            <div className="text-xs text-blue-800 space-y-2">
              <ul className="list-disc list-inside space-y-1 ml-2">
                <li>QQ 邮箱：smtp.qq.com，端口 587 或 465（SSL）</li>
                <li>163 邮箱：smtp.163.com，端口 465（SSL）</li>
                <li>Gmail：smtp.gmail.com，端口 587</li>
                <li>Outlook：smtp-mail.outlook.com，端口 587</li>
              </ul>
              <p className="mt-2 text-amber-700 bg-amber-50 border border-amber-200 rounded px-2 py-1">
                提示：QQ 邮箱和 163 邮箱等需要在邮箱设置中开启 SMTP 服务并使用授权码
              </p>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}
