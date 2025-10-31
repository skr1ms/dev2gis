'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useTheme } from '@/contexts/ThemeContext';
import { useAuth } from '@/hooks/useAuth';

export default function HomePage() {
  const { theme, toggleTheme } = useTheme();
  const { isAuthenticated, loading } = useAuth(false);
  const router = useRouter();

  useEffect(() => {
    if (!loading && isAuthenticated) {
      router.push('/upload');
    }
  }, [isAuthenticated, loading, router]);

  if (loading || isAuthenticated) {
    return null;
  }
  
  return (
    <div className="min-h-screen bg-gradient-surface">
      <nav className="bg-white dark:bg-gray-800 shadow-sm transition-colors">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <svg
                className="w-8 h-8 text-indigo-600 dark:text-indigo-400 mr-2"
                fill="none"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              <span className="text-xl font-bold text-gray-900 dark:text-gray-100">Карты высот БПЛА</span>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={toggleTheme}
                className="p-2 rounded-lg text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                title={theme === 'dark' ? 'Светлая тема' : 'Темная тема'}
              >
                {theme === 'dark' ? (
                  <svg className="w-5 h-5" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                    <path d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                    <path d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                  </svg>
                )}
              </button>
              <Link
                href="/auth/login"
                className="px-6 py-2.5 text-sm font-semibold rounded-lg border-2 border-indigo-600 dark:border-indigo-500 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-600 hover:text-white dark:hover:bg-indigo-500 dark:hover:text-white transition-all duration-300 hover:shadow-lg"
              >
                Войти
              </Link>
            </div>
          </div>
        </div>
      </nav>
      
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="text-center">
          <h1 className="text-5xl md:text-7xl font-extrabold text-gray-900 dark:text-gray-100 mb-6 animate-slideUp">
            <span className="block">Карты высот</span>
            <span className="block text-gradient">из аэрофотосъемки</span>
          </h1>
          <p className="text-xl md:text-2xl text-gray-600 dark:text-gray-400 mb-12 max-w-3xl mx-auto animate-slideUp stagger-delay-1">
            Автоматическая обработка данных БПЛА для создания профессиональных топографических карт высот
          </p>

          <div className="flex flex-col sm:flex-row justify-center gap-4 mb-20 animate-slideUp stagger-delay-2">
            <Link
              href="/auth/register"
              className="px-10 py-4 text-lg font-semibold rounded-lg text-white bg-gradient-primary hover:shadow-xl hover:-translate-y-1 transition-all duration-300"
            >
              Начать работу
            </Link>
            <Link
              href="/auth/login"
              className="px-10 py-4 text-lg font-semibold rounded-lg text-indigo-600 bg-white border-2 border-indigo-600 hover:bg-indigo-50 hover:shadow-xl hover:-translate-y-1 transition-all duration-300"
            >
              Войти в систему
            </Link>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-20">
            <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 hover-lift animate-slideUp stagger-delay-1 transition-colors">
              <div className="flex items-center justify-center w-16 h-16 bg-indigo-100 dark:bg-indigo-900 rounded-full mb-6 mx-auto transition-colors">
                <svg className="w-8 h-8 text-indigo-600 dark:text-indigo-400" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                  <path d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-3">Быстрая обработка</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Одиночное изображение обрабатывается за 2-3 секунды с использованием технологии MiDaS
              </p>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 hover-lift animate-slideUp stagger-delay-2 transition-colors">
              <div className="flex items-center justify-center w-16 h-16 bg-green-100 dark:bg-green-900 rounded-full mb-6 mx-auto transition-colors">
                <svg className="w-8 h-8 text-green-600 dark:text-green-400" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                  <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-3">Точные результаты</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Пакетная фотограмметрия с OpenDroneMap для создания высокоточных карт высот
              </p>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 hover-lift animate-slideUp stagger-delay-3 transition-colors">
              <div className="flex items-center justify-center w-16 h-16 bg-purple-100 dark:bg-purple-900 rounded-full mb-6 mx-auto transition-colors">
                <svg className="w-8 h-8 text-purple-600 dark:text-purple-400" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                  <path d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-3">Красивая визуализация</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Автоматическая генерация отмывки рельефа с цветовым градиентом для анализа местности
              </p>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl p-12 mb-20 animate-slideUp transition-colors">
            <h2 className="text-4xl font-bold text-gray-900 dark:text-gray-100 mb-12">Как это работает</h2>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
              <div className="relative">
                <div className="flex items-center justify-center w-16 h-16 bg-indigo-600 dark:bg-indigo-500 text-white rounded-full mb-4 mx-auto text-2xl font-bold transition-colors">
                  1
                </div>
                <h4 className="font-bold text-lg text-gray-900 dark:text-gray-100 mb-3">Загрузите изображения</h4>
                <p className="text-gray-600 dark:text-gray-400">
                  Одно изображение для быстрой обработки или 5-20 изображений для точной фотограмметрии
                </p>
              </div>
              
              <div className="relative">
                <div className="flex items-center justify-center w-16 h-16 bg-indigo-600 dark:bg-indigo-500 text-white rounded-full mb-4 mx-auto text-2xl font-bold transition-colors">
                  2
                </div>
                <h4 className="font-bold text-lg text-gray-900 dark:text-gray-100 mb-3">Автоматическая обработка</h4>
                <p className="text-gray-600 dark:text-gray-400">
                  Система выбирает оптимальный метод: ИИ-оценка глубины или фотограмметрия
                </p>
              </div>
              
              <div className="relative">
                <div className="flex items-center justify-center w-16 h-16 bg-indigo-600 dark:bg-indigo-500 text-white rounded-full mb-4 mx-auto text-2xl font-bold transition-colors">
                  3
                </div>
                <h4 className="font-bold text-lg text-gray-900 dark:text-gray-100 mb-3">Визуализация</h4>
                <p className="text-gray-600 dark:text-gray-400">
                  Создание отмывки рельефа с цветовым кодированием высот и градиентами
                </p>
              </div>
              
              <div className="relative">
                <div className="flex items-center justify-center w-16 h-16 bg-indigo-600 dark:bg-indigo-500 text-white rounded-full mb-4 mx-auto text-2xl font-bold transition-colors">
                  4
                </div>
                <h4 className="font-bold text-lg text-gray-900 dark:text-gray-100 mb-3">Скачивание</h4>
                <p className="text-gray-600 dark:text-gray-400">
                  Получите готовые карты высот в высоком разрешении в формате PNG
                </p>
              </div>
            </div>
          </div>
          
          <div className="bg-gradient-primary rounded-2xl shadow-2xl p-12 text-white mb-20 animate-slideUp">
            <h2 className="text-3xl font-bold mb-4">Готовы начать?</h2>
            <p className="text-lg mb-8 opacity-90">
              Создайте аккаунт и начните обработку аэрофотоснимков прямо сейчас
            </p>
            <Link
              href="/auth/register"
              className="inline-block px-10 py-4 text-lg font-semibold rounded-lg text-indigo-600 dark:text-indigo-700 bg-white hover:shadow-2xl hover:-translate-y-1 transition-all duration-300"
            >
              Зарегистрироваться бесплатно
            </Link>
          </div>

          <div className="text-gray-600 dark:text-gray-400">
            <p className="text-sm">© 2025 Карты высот БПЛА. Все права защищены.</p>
          </div>
        </div>
      </div>
    </div>
  );
}
