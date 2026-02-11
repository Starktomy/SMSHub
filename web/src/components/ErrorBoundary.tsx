import { Component, type ErrorInfo, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { AlertTriangle, RefreshCcw } from "lucide-react";

interface Props {
  children?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("Uncaught error:", error, errorInfo);
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50 p-4">
          <div className="bg-white p-8 rounded-xl shadow-sm border border-gray-200 max-w-md w-full text-center">
            <div className="w-16 h-16 bg-red-50 rounded-full flex items-center justify-center mx-auto mb-4">
              <AlertTriangle className="w-8 h-8 text-red-500" />
            </div>
            <h1 className="text-xl font-bold text-gray-900 mb-2">出错了</h1>
            <p className="text-gray-500 mb-6 text-sm">
              应用程序遇到了一些问题。请尝试刷新页面。
            </p>
            {this.state.error && (
              <div className="bg-gray-50 p-3 rounded-lg text-left mb-6 overflow-auto max-h-32">
                <code className="text-xs text-red-600 font-mono">
                  {this.state.error.message}
                </code>
              </div>
            )}
            <div className="flex gap-3">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => window.location.href = '/'}
              >
                返回首页
              </Button>
              <Button
                className="flex-1"
                onClick={() => window.location.reload()}
              >
                <RefreshCcw className="w-4 h-4 mr-2" />
                刷新页面
              </Button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
