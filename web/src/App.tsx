import { Suspense, lazy } from 'react';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { QueryProvider } from './providers/QueryProvider';
import { ProtectedRoute } from './components/ProtectedRoute';
import { ErrorBoundary } from './components/ErrorBoundary';
import Layout from './components/Layout';
import { Toaster } from "@/components/ui/sonner.tsx";

// 路由懒加载
const Login = lazy(() => import('./pages/Login'));
const OIDCCallback = lazy(() => import('./pages/OIDCCallback'));
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Messages = lazy(() => import('./pages/Messages'));
const SerialControl = lazy(() => import('./pages/SerialControl'));
const NotificationChannels = lazy(() => import('./pages/NotificationChannels'));
const ScheduledTasksConfig = lazy(() => import('./pages/ScheduledTasksConfig'));
const Devices = lazy(() => import('./pages/Devices'));
const BatchSend = lazy(() => import('./pages/BatchSend'));
const NotFound = lazy(() => import('./pages/NotFound'));

// 加载状态组件
const PageLoader = () => (
  <div className="flex justify-center items-center h-screen bg-gray-50">
    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
  </div>
);

function App() {
    return (
        <ErrorBoundary>
            <QueryProvider>
                <BrowserRouter>
                    <Suspense fallback={<PageLoader />}>
                        <Routes>
                            {/* 公开路由 */}
                            <Route path="/login" element={<Login/>}/>
                            <Route path="/oidc/callback" element={<OIDCCallback/>}/>

                            {/* 受保护的路由 */}
                            <Route
                                path="/"
                                element={
                                    <ProtectedRoute>
                                        <Layout/>
                                    </ProtectedRoute>
                                }
                            >
                                <Route index element={<Dashboard/>}/>
                                <Route path="messages" element={<Messages/>}/>
                                <Route path="devices" element={<Devices/>}/>
                                <Route path="serial" element={<SerialControl/>}/>
                                <Route path="batch-send" element={<BatchSend/>}/>
                                <Route path="notifications" element={<NotificationChannels/>}/>
                                <Route path="scheduled-tasks" element={<ScheduledTasksConfig/>}/>
                            </Route>

                            {/* 404 页面 */}
                            <Route path="*" element={<NotFound />} />
                        </Routes>
                    </Suspense>
                </BrowserRouter>

                <Toaster/>
            </QueryProvider>
        </ErrorBoundary>
    );
}

export default App;
