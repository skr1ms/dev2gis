'use client';

import React from 'react';
import { Badge, getStatusBadgeVariant, getStatusLabel } from '@/components/ui/Badge';
import { formatDate, formatDuration, formatImageCount } from '@/utils/formatters';

interface Heightmap {
  id: string;
  status: string;
  result_url?: string;
  orthophoto_url?: string;
  width?: number;
  height?: number;
  processing_time?: number;
  created_at: string;
  image_count?: number;
  merge_method?: string;
}

interface ResultCardProps {
  heightmap: Heightmap;
  type: 'single' | 'batch';
}

export function ResultCard({ heightmap, type }: ResultCardProps) {
  const hasHeightmap = heightmap.status === 'completed' && heightmap.result_url;
  const hasOrthophoto = heightmap.status === 'completed' && heightmap.orthophoto_url;
  
  const handleView = (url: string) => {
    window.open(url, '_blank');
  };
  
  const handleDownload = async (url: string, filename: string) => {
    try {
      const response = await fetch(url);
      if (!response.ok) throw new Error('Failed to fetch file');
      
      const blob = await response.blob();
      const blobUrl = window.URL.createObjectURL(blob);
      
      const a = document.createElement('a');
      a.href = blobUrl;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      
      document.body.removeChild(a);
      window.URL.revokeObjectURL(blobUrl);
    } catch (error) {
      console.error('Failed to download:', error);
      alert('Не удалось скачать файл');
    }
  };
  
  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg hover:shadow-xl transition-all duration-300 overflow-hidden">
      <div className="p-6 flex flex-col h-full">
          <div className="flex items-center justify-between mb-3">
            <Badge variant={getStatusBadgeVariant(heightmap.status)}>
              {getStatusLabel(heightmap.status)}
            </Badge>
            {type === 'batch' && (
              <span className="text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">
                Пакетная
              </span>
            )}
          </div>
          
          <div className="space-y-2 text-sm text-gray-600 dark:text-gray-300 flex-grow">
            <div className="flex items-center">
              <svg className="w-4 h-4 mr-2 text-gray-400 dark:text-gray-500" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clipRule="evenodd" />
              </svg>
              <span>{formatDate(heightmap.created_at)}</span>
            </div>
            
            {heightmap.width && heightmap.height && (
              <div className="flex items-center">
                <svg className="w-4 h-4 mr-2 text-gray-400 dark:text-gray-500" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 12v3c0 1.657 3.134 3 7 3s7-1.343 7-3v-3c0 1.657-3.134 3-7 3s-7-1.343-7-3z" />
                  <path d="M3 7v3c0 1.657 3.134 3 7 3s7-1.343 7-3V7c0 1.657-3.134 3-7 3S3 8.657 3 7z" />
                  <path d="M17 5c0 1.657-3.134 3-7 3S3 6.657 3 5s3.134-3 7-3 7 1.343 7 3z" />
                </svg>
                <span>{heightmap.width} × {heightmap.height} px</span>
              </div>
            )}
            
            {heightmap.image_count && (
              <div className="flex items-center">
                <svg className="w-4 h-4 mr-2 text-gray-400 dark:text-gray-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
                </svg>
                <span>{formatImageCount(heightmap.image_count)}</span>
              </div>
            )}
            
            {heightmap.processing_time && (
              <div className="flex items-center">
                <svg className="w-4 h-4 mr-2 text-gray-400 dark:text-gray-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
                </svg>
                <span>{formatDuration(heightmap.processing_time)}</span>
              </div>
            )}
          </div>
          
          {(hasHeightmap || hasOrthophoto) && (
            <div className="mt-4 space-y-3">
              {hasHeightmap && (
                <div className="space-y-2">
                  {hasOrthophoto && (
                    <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Карта высот (DSM)
                    </div>
                  )}
                  <div className="grid grid-cols-2 gap-3">
                    <button
                      onClick={() => handleView(heightmap.result_url!)}
                      className="h-10 px-4 bg-green-600 dark:bg-green-500 text-white rounded-lg hover:bg-green-700 dark:hover:bg-green-600 transition-all duration-200 text-sm font-medium flex items-center justify-center gap-2 shadow-sm hover:shadow-md"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      Просмотр
                    </button>
                    
                    <button
                      onClick={() => handleDownload(heightmap.result_url!, `heightmap_${heightmap.id}.png`)}
                      className="h-10 px-4 bg-indigo-600 dark:bg-indigo-500 text-white rounded-lg hover:bg-indigo-700 dark:hover:bg-indigo-600 transition-all duration-200 text-sm font-medium flex items-center justify-center gap-2 shadow-sm hover:shadow-md"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                      </svg>
                      Скачать
                    </button>
                  </div>
                </div>
              )}
              
              {hasOrthophoto && (
                <div className="space-y-2">
                  {hasHeightmap && (
                    <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Ортофотоплан
                    </div>
                  )}
                  <button
                    onClick={() => handleDownload(heightmap.orthophoto_url!, `orthophoto_${heightmap.id}.tif`)}
                    className="w-full h-10 px-4 bg-orange-600 dark:bg-orange-500 text-white rounded-lg hover:bg-orange-700 dark:hover:bg-orange-600 transition-all duration-200 text-sm font-medium flex items-center justify-center gap-2 shadow-sm hover:shadow-md"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                    </svg>
                    Скачать ортофотоплан (.tif)
                  </button>
                </div>
              )}
            </div>
          )}
      </div>
    </div>
  );
}

