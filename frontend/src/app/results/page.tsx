'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { useToast } from '@/contexts/ToastContext';
import { Navbar } from '@/components/layout/Navbar';
import { Footer } from '@/components/layout/Footer';
import { ResultCard } from '@/components/results/ResultCard';
import { FilterBar } from '@/components/results/FilterBar';
import { Loading } from '@/components/ui/Spinner';
import { SkeletonGrid } from '@/components/ui/Skeleton';
import { api } from '@/utils/api';

interface Heightmap {
  id: string;
  status: string;
  result_url?: string;
  width?: number;
  height?: number;
  processing_time?: number;
  created_at: string;
  image_count?: number;
  merge_method?: string;
}

type SortOption = 'newest' | 'oldest' | 'status';
type StatusFilter = 'all' | 'completed' | 'processing' | 'pending' | 'failed';

export default function ResultsPage() {
  const { user, loading: authLoading } = useAuth(true);
  const { error: showError, info } = useToast();
  
  const [singleMaps, setSingleMaps] = useState<Heightmap[]>([]);
  const [batchMaps, setBatchMaps] = useState<Heightmap[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<SortOption>('newest');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [autoRefresh, setAutoRefresh] = useState(true);

  useEffect(() => {
    if (!authLoading && user) {
      fetchHeightmaps();
      
      if (autoRefresh) {
        const interval = setInterval(fetchHeightmaps, 5000);
        return () => clearInterval(interval);
      }
    }
  }, [authLoading, user, autoRefresh]);

  const fetchHeightmaps = async () => {
    try {
      const [singleResponse, batchResponse] = await Promise.all([
        api.get('/api/heightmaps?limit=100'),
        api.get('/api/heightmaps/batch?limit=100')
      ]);
      
      const single = Array.isArray(singleResponse) ? singleResponse : (singleResponse.heightmaps || []);
      const batch = Array.isArray(batchResponse) ? batchResponse : (batchResponse.batch_heightmaps || []);
      
      setSingleMaps(single);
      setBatchMaps(batch);
    } catch (error: any) {
      showError(error.message || 'Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setLoading(true);
    await fetchHeightmaps();
    info('Данные обновлены');
  };

  if (authLoading) {
    return <Loading text="Загрузка..." />;
  }

  if (!user) {
    return null;
  }

  const allMaps = [
    ...singleMaps.map(m => ({ ...m, type: 'single' as const })),
    ...batchMaps.map(m => ({ ...m, type: 'batch' as const }))
  ];

  const filteredMaps = allMaps
    .filter(map => {
      if (statusFilter !== 'all' && map.status !== statusFilter) return false;
      if (searchQuery && !map.id.toLowerCase().includes(searchQuery.toLowerCase())) return false;
      return true;
    })
    .sort((a, b) => {
      if (sortBy === 'newest') {
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
      }
      if (sortBy === 'oldest') {
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      }
      if (sortBy === 'status') {
        const statusOrder = { completed: 0, processing: 1, pending: 2, failed: 3 };
        return (statusOrder[a.status as keyof typeof statusOrder] || 99) - 
               (statusOrder[b.status as keyof typeof statusOrder] || 99);
      }
      return 0;
    });

  const stats = {
    total: allMaps.length,
    completed: allMaps.filter(m => m.status === 'completed').length,
    processing: allMaps.filter(m => m.status === 'processing').length,
    pending: allMaps.filter(m => m.status === 'pending').length,
    failed: allMaps.filter(m => m.status === 'failed').length,
  };

  return (
    <>
      <Navbar />
      <main className="flex-1 bg-gradient-surface">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex items-center justify-between mb-8 animate-slideUp">
            <div>
              <h1 className="text-4xl font-bold text-gray-900 mb-2">Результаты обработки</h1>
              <p className="text-gray-600">
                Всего: {stats.total} | Завершено: {stats.completed} | В обработке: {stats.processing}
              </p>
            </div>
            
            <div className="flex items-center gap-4">
              <label className="flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={autoRefresh}
                  onChange={(e) => setAutoRefresh(e.target.checked)}
                  className="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
                />
                <span className="ml-2 text-sm text-gray-700">Автообновление</span>
              </label>
              
              <button
                onClick={() => window.location.href = '/upload'}
                className="px-6 py-2 bg-indigo-600 text-white rounded-lg font-medium hover:bg-indigo-700 transition-all duration-300 hover:shadow-lg"
              >
                Новая загрузка
              </button>
            </div>
          </div>

          <FilterBar
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            sortBy={sortBy}
            onSortChange={setSortBy}
            statusFilter={statusFilter}
            onStatusFilterChange={setStatusFilter}
            onRefresh={handleRefresh}
          />

          {loading ? (
            <SkeletonGrid count={6} />
          ) : filteredMaps.length === 0 ? (
            <div className="bg-white rounded-xl shadow-lg p-12 text-center animate-slideUp">
              <svg
                className="w-24 h-24 text-gray-400 mx-auto mb-4"
                fill="none"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
              </svg>
              <h3 className="text-2xl font-bold text-gray-900 mb-2">Результатов не найдено</h3>
              <p className="text-gray-600 mb-6">
                {allMaps.length === 0 
                  ? 'Загрузите изображения для создания карт высот'
                  : 'Попробуйте изменить параметры фильтрации'}
              </p>
              {allMaps.length === 0 && (
                <button
                  onClick={() => window.location.href = '/upload'}
                  className="px-6 py-3 bg-indigo-600 text-white rounded-lg font-medium hover:bg-indigo-700 transition-all duration-300 hover:shadow-lg"
                >
                  Загрузить изображения
                </button>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-slideUp">
              {filteredMaps.map((map) => (
                <ResultCard
                  key={map.id}
                  heightmap={map}
                  type={map.type}
                />
              ))}
            </div>
          )}

          {filteredMaps.length > 0 && filteredMaps.length < allMaps.length && (
            <div className="mt-6 text-center text-gray-600 animate-slideUp">
              <p>Показано {filteredMaps.length} из {allMaps.length} результатов</p>
            </div>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
