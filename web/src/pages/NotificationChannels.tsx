import {useEffect, useState} from 'react';
import {Loader2, Save} from 'lucide-react';
import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query';
import {toast} from 'sonner';
import {Button} from '@/components/ui/button';
import {
    getNotificationChannels,
    type NotificationChannel,
    saveNotificationChannels,
    testNotificationChannel
} from "@/api/property.ts";

import { DingtalkConfig } from '@/components/notification-channels/DingtalkConfig';
import { WecomConfig } from '@/components/notification-channels/WecomConfig';
import { FeishuConfig } from '@/components/notification-channels/FeishuConfig';
import { WebhookConfig } from '@/components/notification-channels/WebhookConfig';
import { EmailConfig } from '@/components/notification-channels/EmailConfig';
import { TelegramConfig } from '@/components/notification-channels/TelegramConfig';

interface FormValues {
    // 钉钉
    dingtalkEnabled: boolean;
    dingtalkSecretKey: string;
    dingtalkSignSecret: string;

    // 企业微信
    wecomEnabled: boolean;
    wecomSecretKey: string;

    // 飞书
    feishuEnabled: boolean;
    feishuSecretKey: string;
    feishuSignSecret: string;

    // Webhook
    webhookEnabled: boolean;
    webhookUrl: string;
    webhookMethod: string;
    webhookContentType: string;
    webhookHeaders: string;
    webhookBody: string;

    // 邮件
    emailEnabled: boolean;
    emailSmtpHost: string;
    emailSmtpPort: string;
    emailUsername: string;
    emailPassword: string;
    emailFrom: string;
    emailTo: string;
    emailSubject: string;

    //telegram
    telegramlEnabled: boolean;
    telegramApiToken: string;
    telegramUserid: string;
    telegramProxyEnabled: boolean;
    telegramProxyUrl: string;
    telegramProxyUsername: string;
    telegramProxyPassword: string;
}

export default function NotificationChannels() {
    const queryClient = useQueryClient();
    const [formValues, setFormValues] = useState<FormValues>({
        dingtalkEnabled: false,
        dingtalkSecretKey: '',
        dingtalkSignSecret: '',
        wecomEnabled: false,
        wecomSecretKey: '',
        feishuEnabled: false,
        feishuSecretKey: '',
        feishuSignSecret: '',
        webhookEnabled: false,
        webhookUrl: '',
        webhookMethod: 'POST',
        webhookContentType: 'application/json; charset=utf-8',
        webhookHeaders: '',
        webhookBody: '{"from": "{{from}}", "content": "{{content}}", "timestamp": "{{timestamp}}"}',
        emailEnabled: false,
        emailSmtpHost: '',
        emailSmtpPort: '587',
        emailUsername: '',
        emailPassword: '',
        emailFrom: '',
        emailTo: '',
        emailSubject: '收到新短信 - {{from}}',
        telegramlEnabled: false,
        telegramApiToken: '',
        telegramUserid: '',
        telegramProxyEnabled: false,
        telegramProxyUrl: '',
        telegramProxyUsername: '',
        telegramProxyPassword: '',
    });

    // 获取通知渠道列表
    const {data: channels = [], isLoading} = useQuery({
        queryKey: ['notificationChannels'],
        queryFn: getNotificationChannels,
    });

    // 保存 mutation
    const saveMutation = useMutation({
        mutationFn: saveNotificationChannels,
        onSuccess: () => {
            toast.success('保存成功');
            queryClient.invalidateQueries({queryKey: ['notificationChannels']});
        },
        onError: (error: unknown) => {
            console.error('保存失败:', error);
            toast.error('保存失败');
        },
    });

    // 测试 mutation
    const testMutation = useMutation({
        mutationFn: testNotificationChannel,
        onSuccess: () => {
            toast.success('测试通知已发送，请检查对应渠道');
        },
        onError: (error: unknown) => {
            console.error('测试失败:', error);
            toast.error('测试失败，请检查配置');
        },
    });

    // 将渠道数组转换为表单值
    useEffect(() => {
        if (channels.length > 0) {
            const newFormValues: FormValues = {...formValues};

            channels.forEach((channel) => {
                if (channel.type === 'dingtalk') {
                    newFormValues.dingtalkEnabled = channel.enabled;
                    newFormValues.dingtalkSecretKey = (channel.config?.secretKey as string) || '';
                    newFormValues.dingtalkSignSecret = (channel.config?.signSecret as string) || '';
                } else if (channel.type === 'wecom') {
                    newFormValues.wecomEnabled = channel.enabled;
                    newFormValues.wecomSecretKey = (channel.config?.secretKey as string) || '';
                } else if (channel.type === 'feishu') {
                    newFormValues.feishuEnabled = channel.enabled;
                    newFormValues.feishuSecretKey = (channel.config?.secretKey as string) || '';
                    newFormValues.feishuSignSecret = (channel.config?.signSecret as string) || '';
                } else if (channel.type === 'webhook') {
                    newFormValues.webhookEnabled = channel.enabled;
                    newFormValues.webhookUrl = (channel.config?.url as string) || '';
                    newFormValues.webhookMethod = (channel.config?.method as string) || 'POST';
                    newFormValues.webhookContentType = (channel.config?.contentType as string) || 'application/json; charset=utf-8';
                    newFormValues.webhookBody = (channel.config?.body as string) || '{"from": "{{from}}", "content": "{{content}}", "timestamp": "{{timestamp}}"}';

                    // 解析 headers 为 JSON 字符串
                    const headers = channel.config?.headers || {};
                    newFormValues.webhookHeaders = JSON.stringify(headers, null, 2);
                } else if (channel.type === 'email') {
                    newFormValues.emailEnabled = channel.enabled;
                    newFormValues.emailSmtpHost = (channel.config?.smtpHost as string) || '';
                    newFormValues.emailSmtpPort = (channel.config?.smtpPort as string) || '587';
                    newFormValues.emailUsername = (channel.config?.username as string) || '';
                    newFormValues.emailPassword = (channel.config?.password as string) || '';
                    newFormValues.emailFrom = (channel.config?.from as string) || '';
                    newFormValues.emailTo = (channel.config?.to as string) || '';
                    newFormValues.emailSubject = (channel.config?.subject as string) || '收到新短信 - {{from}}';
                } else if (channel.type === 'telegram') {
                    newFormValues.telegramlEnabled = channel.enabled;
                    newFormValues.telegramApiToken = (channel.config?.apiToken as string) || '';
                    newFormValues.telegramUserid = (channel.config?.userid as string) || '';
                    newFormValues.telegramProxyEnabled = (channel.config?.proxyEnabled as boolean)||false;
                    newFormValues.telegramProxyUrl = (channel.config?.proxyUrl as string) || '';
                    newFormValues.telegramProxyUsername = (channel.config?.proxyUsername as string) || '';
                    newFormValues.telegramProxyPassword = (channel.config?.proxyPassword as string) || '';
                }
            });

            setFormValues(newFormValues);
        }
    }, [channels]);

    // 更新表单字段
    const updateField = (field: keyof FormValues, value: any) => {
        setFormValues((prev) => ({...prev, [field]: value}));
    };

    // 保存配置
    const handleSave = async () => {
        const newChannels: NotificationChannel[] = [];

        // 钉钉
        if (formValues.dingtalkEnabled || formValues.dingtalkSecretKey) {
            newChannels.push({
                type: 'dingtalk',
                enabled: formValues.dingtalkEnabled,
                config: {
                    secretKey: formValues.dingtalkSecretKey,
                    signSecret: formValues.dingtalkSignSecret,
                },
            });
        }

        // 企业微信
        if (formValues.wecomEnabled || formValues.wecomSecretKey) {
            newChannels.push({
                type: 'wecom',
                enabled: formValues.wecomEnabled,
                config: {
                    secretKey: formValues.wecomSecretKey,
                },
            });
        }

        // 飞书
        if (formValues.feishuEnabled || formValues.feishuSecretKey) {
            newChannels.push({
                type: 'feishu',
                enabled: formValues.feishuEnabled,
                config: {
                    secretKey: formValues.feishuSecretKey,
                    signSecret: formValues.feishuSignSecret,
                },
            });
        }

        // Webhook
        if (formValues.webhookEnabled || formValues.webhookUrl) {
            let headers = {};
            if (formValues.webhookHeaders) {
                try {
                    headers = JSON.parse(formValues.webhookHeaders);
                } catch (err) {
                    toast.error('Webhook Headers JSON 格式错误');
                    return;
                }
            }

            newChannels.push({
                type: 'webhook',
                enabled: formValues.webhookEnabled,
                config: {
                    url: formValues.webhookUrl,
                    method: formValues.webhookMethod,
                    contentType: formValues.webhookContentType,
                    body: formValues.webhookBody,
                    headers: Object.keys(headers).length > 0 ? headers : undefined,
                },
            });
        }

        // 邮件
        if (formValues.emailEnabled || formValues.emailSmtpHost) {
            newChannels.push({
                type: 'email',
                enabled: formValues.emailEnabled,
                config: {
                    smtpHost: formValues.emailSmtpHost,
                    smtpPort: formValues.emailSmtpPort,
                    username: formValues.emailUsername,
                    password: formValues.emailPassword,
                    from: formValues.emailFrom,
                    to: formValues.emailTo,
                    subject: formValues.emailSubject,
                },
            });
        }

        if (formValues.telegramlEnabled||formValues.telegramApiToken) {
            if (formValues.telegramProxyEnabled && !formValues.telegramProxyUrl) {
                toast.error('已启用 HTTP 代理，但未填写代理地址')
                return
            }

            newChannels.push({
                type:'telegram',
                enabled:formValues.telegramlEnabled,
                config: {
                    apiToken: formValues.telegramApiToken,
                    userid: formValues.telegramUserid,
                    proxyEnabled: formValues.telegramProxyEnabled,
                    proxyUrl: formValues.telegramProxyUrl,
                    proxyUsername: formValues.telegramProxyUsername,
                    proxyPassword: formValues.telegramProxyPassword,
                }
            })
        }

        saveMutation.mutate(newChannels);
    };

    if (isLoading) {
        return (
            <div className="flex justify-center items-center py-20">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <div className="space-y-8 animate-in fade-in duration-300">
            <div className="border-b border-gray-200 pb-5">
                <h1 className="text-2xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent flex items-center gap-3">
                    通知渠道管理
                </h1>
                <p className="text-sm text-gray-500 mt-3">配置第三方消息推送渠道，当收到短信或设备异常时自动推送通知</p>
            </div>

            <div className="grid grid-cols-1 gap-6">
                {/* 钉钉通知 */}
                <DingtalkConfig
                    enabled={formValues.dingtalkEnabled}
                    secretKey={formValues.dingtalkSecretKey}
                    signSecret={formValues.dingtalkSignSecret}
                    onUpdate={updateField}
                    onTest={() => testMutation.mutate('dingtalk')}
                    isTestPending={testMutation.isPending}
                />

                {/* 企业微信通知 */}
                <WecomConfig
                    enabled={formValues.wecomEnabled}
                    secretKey={formValues.wecomSecretKey}
                    onUpdate={updateField}
                    onTest={() => testMutation.mutate('wecom')}
                    isTestPending={testMutation.isPending}
                />

                {/* 飞书通知 */}
                <FeishuConfig
                    enabled={formValues.feishuEnabled}
                    secretKey={formValues.feishuSecretKey}
                    signSecret={formValues.feishuSignSecret}
                    onUpdate={updateField}
                    onTest={() => testMutation.mutate('feishu')}
                    isTestPending={testMutation.isPending}
                />

                {/* 自定义 Webhook */}
                <WebhookConfig
                    enabled={formValues.webhookEnabled}
                    url={formValues.webhookUrl}
                    method={formValues.webhookMethod}
                    contentType={formValues.webhookContentType}
                    headers={formValues.webhookHeaders}
                    body={formValues.webhookBody}
                    onUpdate={updateField}
                    onTest={() => testMutation.mutate('webhook')}
                    isTestPending={testMutation.isPending}
                />

                {/* 邮件通知 */}
                <EmailConfig
                    enabled={formValues.emailEnabled}
                    smtpHost={formValues.emailSmtpHost}
                    smtpPort={formValues.emailSmtpPort}
                    username={formValues.emailUsername}
                    password={formValues.emailPassword}
                    from={formValues.emailFrom}
                    to={formValues.emailTo}
                    subject={formValues.emailSubject}
                    onUpdate={updateField}
                    onTest={() => testMutation.mutate('email')}
                    isTestPending={testMutation.isPending}
                />

                {/* telegram通知 */}
                <TelegramConfig
                    enabled={formValues.telegramlEnabled}
                    apiToken={formValues.telegramApiToken}
                    userid={formValues.telegramUserid}
                    proxyEnabled={formValues.telegramProxyEnabled}
                    proxyUrl={formValues.telegramProxyUrl}
                    proxyUsername={formValues.telegramProxyUsername}
                    proxyPassword={formValues.telegramProxyPassword}
                    onUpdate={updateField}
                    onTest={() => testMutation.mutate('telegram')}
                    isTestPending={testMutation.isPending}
                />

                {/* 保存按钮 */}
                <div className="flex pt-6 border-t border-gray-200">
                    <Button
                        onClick={handleSave}
                        disabled={saveMutation.isPending}
                        className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 transition-all px-8 py-2.5 text-sm font-medium min-w-[140px]"
                    >
                        {saveMutation.isPending ? (
                            <>
                                <Loader2 className="w-4 h-4 mr-2 animate-spin"/>
                                保存中...
                            </>
                        ) : (
                            <>
                                <Save className="w-4 h-4 mr-2"/>
                                保存配置
                            </>
                        )}
                    </Button>
                </div>
            </div>
        </div>
    );
}
