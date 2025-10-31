'use client';

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Navbar } from '@/components/layout/Navbar';
import { Footer } from '@/components/layout/Footer';
import { StatsCard, ServiceLinkCard } from '@/components/admin/StatsCard';
import { Loading } from '@/components/ui/Spinner';
import { SkeletonGrid } from '@/components/ui/Skeleton';
import { api } from '@/utils/api';
import { formatDate } from '@/utils/formatters';

interface AdminStats {
  totalUsers: number;
  totalHeightmaps: number;
  totalBatchHeightmaps: number;
  statusCounts: {
    completed: number;
    processing: number;
    pending: number;
    failed: number;
  };
}

interface RecentHeightmap {
  id: string;
  type: 'single' | 'batch';
  status: string;
  created_at: string;
  user_email?: string;
}

export default function AdminPage() {
  const { user, loading: authLoading, isAdmin } = useAuth(true, true);
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [recentMaps, setRecentMaps] = useState<RecentHeightmap[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  useEffect(() => {
    if (!authLoading && isAdmin) {
      fetchAdminData();
    }
  }, [authLoading, isAdmin]);
  
  const fetchAdminData = async () => {
    try {
      setLoading(true);
      
      const singleMaps = await api.get('/api/heightmaps?limit=100');
      const batchMaps = await api.get('/api/heightmaps/batch?limit=100');
      
      const allMaps = [
        ...(Array.isArray(singleMaps) ? singleMaps : singleMaps.heightmaps || []),
        ...(Array.isArray(batchMaps) ? batchMaps : batchMaps.batch_heightmaps || [])
      ];
      
      const statsData: AdminStats = {
        totalUsers: 1,
        totalHeightmaps: (Array.isArray(singleMaps) ? singleMaps : singleMaps.heightmaps || []).length,
        totalBatchHeightmaps: (Array.isArray(batchMaps) ? batchMaps : batchMaps.batch_heightmaps || []).length,
        statusCounts: {
          completed: allMaps.filter((m: any) => m.status === 'completed').length,
          processing: allMaps.filter((m: any) => m.status === 'processing').length,
          pending: allMaps.filter((m: any) => m.status === 'pending').length,
          failed: allMaps.filter((m: any) => m.status === 'failed').length,
        },
      };
      
      setStats(statsData);
      
      const recent = allMaps
        .sort((a: any, b: any) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
        .slice(0, 10)
        .map((m: any): RecentHeightmap => ({
          id: m.id,
          type: m.image_count ? 'batch' : 'single',
          status: m.status,
          created_at: m.created_at,
        }));
      
      setRecentMaps(recent);
    } catch (err: any) {
      setError(err.message || 'Не удалось загрузить данные');
    } finally {
      setLoading(false);
    }
  };
  
  if (authLoading) {
    return <Loading text="Проверка прав доступа..." />;
  }
  
  if (!user || !isAdmin) {
    return null;
  }
  
  return (
    <>
      <Navbar />
      <main className="flex-1 bg-gradient-surface">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="mb-8 animate-slideUp">
            <h1 className="text-4xl font-bold text-gray-900 mb-2">Админ панель</h1>
            <p className="text-gray-600">Мониторинг системы и управление сервисами</p>
          </div>
          
          {error && (
            <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg animate-slideDown">
              {error}
            </div>
          )}
          
          {loading ? (
            <SkeletonGrid count={4} />
          ) : (
            <>
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-4">Статистика системы</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  <div className="animate-slideUp stagger-delay-1">
                    <StatsCard
                      title="Всего пользователей"
                      value={stats?.totalUsers || 0}
                      color="blue"
                      icon={
                        <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z" />
                        </svg>
                      }
                    />
                  </div>
                  
                  <div className="animate-slideUp stagger-delay-2">
                    <StatsCard
                      title="Одиночные карты"
                      value={stats?.totalHeightmaps || 0}
                      color="purple"
                      icon={
                        <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
                        </svg>
                      }
                    />
                  </div>
                  
                  <div className="animate-slideUp stagger-delay-3">
                    <StatsCard
                      title="Пакетные карты"
                      value={stats?.totalBatchHeightmaps || 0}
                      color="green"
                      icon={
                        <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M3 12v3c0 1.657 3.134 3 7 3s7-1.343 7-3v-3c0 1.657-3.134 3-7 3s-7-1.343-7-3z" />
                          <path d="M3 7v3c0 1.657 3.134 3 7 3s7-1.343 7-3V7c0 1.657-3.134 3-7 3S3 8.657 3 7z" />
                          <path d="M17 5c0 1.657-3.134 3-7 3S3 6.657 3 5s3.134-3 7-3 7 1.343 7 3z" />
                        </svg>
                      }
                    />
                  </div>
                  
                  <div className="animate-slideUp stagger-delay-4">
                    <StatsCard
                      title="Всего задач"
                      value={(stats?.totalHeightmaps || 0) + (stats?.totalBatchHeightmaps || 0)}
                      color="indigo"
                      icon={
                        <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
                          <path fillRule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clipRule="evenodd" />
                        </svg>
                      }
                    />
                  </div>
                </div>
              </div>
              
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-4">Статусы задач</h2>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <StatsCard
                    title="Завершено"
                    value={stats?.statusCounts.completed || 0}
                    color="green"
                  />
                  <StatsCard
                    title="В обработке"
                    value={stats?.statusCounts.processing || 0}
                    color="yellow"
                  />
                  <StatsCard
                    title="В ожидании"
                    value={stats?.statusCounts.pending || 0}
                    color="blue"
                  />
                  <StatsCard
                    title="Ошибки"
                    value={stats?.statusCounts.failed || 0}
                    color="red"
                  />
                </div>
              </div>
              
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-4">Внешние сервисы</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <ServiceLinkCard
                    title="Grafana"
                    description="Мониторинг и метрики системы"
                    url="http://localhost:3000"
                    color="indigo"
                    icon={
                      <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
                      </svg>
                    }
                  />
                  
                  <ServiceLinkCard
                    title="MinIO Console"
                    description="Управление хранилищем S3"
                    url="http://localhost:9001"
                    color="green"
                    icon={
                      <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M3 12v3c0 1.657 3.134 3 7 3s7-1.343 7-3v-3c0 1.657-3.134 3-7 3s-7-1.343-7-3z" />
                        <path d="M3 7v3c0 1.657 3.134 3 7 3s7-1.343 7-3V7c0 1.657-3.134 3-7 3S3 8.657 3 7z" />
                        <path d="M17 5c0 1.657-3.134 3-7 3S3 6.657 3 5s3.134-3 7-3 7 1.343 7 3z" />
                      </svg>
                    }
                  />
                </div>
              </div>
              
              {recentMaps.length > 0 && (
                <div>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Последние задачи</h2>
                  <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            ID
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Тип
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Статус
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Создано
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {recentMaps.map((map) => (
                          <tr key={map.id} className="hover:bg-gray-50 transition-colors">
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                              {map.id.slice(0, 8)}...
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {map.type === 'batch' ? 'Пакет' : 'Одиночное'}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span
                                className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${
                                  map.status === 'completed'
                                    ? 'bg-green-100 text-green-800'
                                    : map.status === 'failed'
                                    ? 'bg-red-100 text-red-800'
                                    : map.status === 'processing'
                                    ? 'bg-yellow-100 text-yellow-800'
                                    : 'bg-blue-100 text-blue-800'
                                }`}
                              >
                                {map.status === 'completed' ? 'Завершено' :
                                 map.status === 'failed' ? 'Ошибка' :
                                 map.status === 'processing' ? 'Обработка' : 'Ожидание'}
                              </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                              {formatDate(map.created_at)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}

