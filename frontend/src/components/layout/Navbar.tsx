'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { ThemeSelector } from '@/components/ui/ThemeSelector';

export function Navbar() {
  const pathname = usePathname();
  const { user, isAuthenticated, isAdmin, logout } = useAuth();
  
  const isActive = (path: string) => pathname === path;
  
  const linkClasses = (path: string) =>
    `px-3 py-2 rounded-md text-sm font-medium transition-colors ${
      isActive(path)
        ? 'bg-indigo-700 text-white'
        : 'text-indigo-100 hover:bg-indigo-600 hover:text-white'
    }`;
  
  if (!isAuthenticated) {
    return null;
  }
  
  return (
    <nav className="bg-indigo-600 dark:bg-gray-800 shadow-lg transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
            <Link href="/" className="flex items-center hover:opacity-80 transition-opacity">
              <svg
                className="w-8 h-8 text-white mr-2"
                fill="none"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              <span className="text-xl font-bold text-white">Карты высот БПЛА</span>
            </Link>
          </div>
          
          <div className="flex items-center space-x-4">
            <Link href="/upload" className={linkClasses('/upload')}>
              Загрузка
            </Link>
            <Link href="/results" className={linkClasses('/results')}>
              Результаты
            </Link>
            {isAdmin && (
              <Link href="/admin" className={linkClasses('/admin')}>
                Админ
              </Link>
            )}
            
            <div className="ml-4 flex items-center space-x-4">
              <ThemeSelector />
              
              <div className="text-sm text-indigo-100">
                <span className="font-medium">{user?.name}</span>
                {isAdmin && (
                  <span className="ml-2 px-2 py-1 bg-indigo-800 rounded text-xs">
                    Админ
                  </span>
                )}
              </div>
              <button
                onClick={logout}
                className="px-3 py-2 rounded-md text-sm font-medium text-indigo-100 hover:bg-indigo-700 hover:text-white transition-colors"
              >
                Выход
              </button>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
}

